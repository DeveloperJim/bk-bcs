/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package multirelease

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/internal/audit"
	"bk-bscp/internal/authorization"
	"bk-bscp/internal/database"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbbcscontroller "bk-bscp/internal/protocol/bcs-controller"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// PublishAction publishes target multi release object.
type PublishAction struct {
	kit              kit.Kit
	viper            *viper.Viper
	authSvrCli       pbauthserver.AuthClient
	dataMgrCli       pbdatamanager.DataManagerClient
	bcsControllerCli pbbcscontroller.BCSControllerClient
	gseControllerCli pbgsecontroller.GSEControllerClient

	req  *pb.PublishMultiReleaseReq
	resp *pb.PublishMultiReleaseResp

	multiRelease *pbcommon.MultiRelease
	app          *pbcommon.App

	releaseIDs []string
	commitIDs  []string
}

// NewPublishAction creates new PublishAction.
func NewPublishAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	bcsControllerCli pbbcscontroller.BCSControllerClient, gseControllerCli pbgsecontroller.GSEControllerClient,
	req *pb.PublishMultiReleaseReq, resp *pb.PublishMultiReleaseResp) *PublishAction {

	action := &PublishAction{
		kit:              kit,
		viper:            viper,
		authSvrCli:       authSvrCli,
		dataMgrCli:       dataMgrCli,
		bcsControllerCli: bcsControllerCli,
		gseControllerCli: gseControllerCli,
		req:              req,
		resp:             resp,
	}

	action.resp.Result = true
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PublishAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PublishAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Authorize checks the action authorization.
func (act *PublishAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *PublishAction) Output() error {
	// do nothing.
	return nil
}

func (act *PublishAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_release_id", act.req.MultiReleaseId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *PublishAction) authorize() (pbcommon.ErrCode, string) {
	isAuthorized, err := authorization.Authorize(act.kit, act.req.AppId, auth.LocalAuthAction,
		act.authSvrCli, act.viper.GetDuration("authserver.callTimeout"))
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("authorize failed, %+v", err)
	}

	if !isAuthorized {
		return pbcommon.ErrCode_E_NOT_AUTHORIZED, "not authorized"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) queryApp() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAppReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		AppId: act.multiRelease.AppId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	act.app = resp.Data

	return resp.Code, resp.Message
}

func (act *PublishAction) querySubReleaseList() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseSubListReq{
		Seq:            act.kit.Rid,
		BizId:          act.req.BizId,
		MultiReleaseId: act.req.MultiReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiReleaseSubList(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to datamanager QueryMultiReleaseSubList, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.releaseIDs = resp.Data.ReleaseIds

	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) querySubCommitList() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiCommitSubListReq{
		Seq:           act.kit.Rid,
		BizId:         act.req.BizId,
		MultiCommitId: act.multiRelease.MultiCommitId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiCommitSubList(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to datamanager QueryMultiCommitSubList, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.commitIDs = resp.Data.CommitIds

	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) queryMultiRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseReq{
		Seq:            act.kit.Rid,
		BizId:          act.req.BizId,
		MultiReleaseId: act.req.MultiReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryMultiRelease, %+v", err)
	}
	act.multiRelease = resp.Data

	return resp.Code, resp.Message
}

func (act *PublishAction) publishPreBCSMode(releaseID string) (pbcommon.ErrCode, string) {
	r := &pbbcscontroller.PublishReleasePreReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: releaseID,
		Operator:  act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("bcscontroller.callTimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%s]| request to bcs-controller, %+v", r.Seq, r)

	resp, err := act.bcsControllerCli.PublishReleasePre(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to bcs-controller PublishReleasePre, %+v", err)
	}

	if resp.Code == pbcommon.ErrCode_E_BCS_ALREADY_PUBLISHED {
		return pbcommon.ErrCode_E_OK, ""
	}
	return resp.Code, resp.Message
}

func (act *PublishAction) publishPreGSEPluginMode(releaseID string) (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.PublishReleasePreReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: releaseID,
		Operator:  act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("gsecontroller.callTimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%s]| request to gse-controller, %+v", r.Seq, r)

	resp, err := act.gseControllerCli.PublishReleasePre(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to gse-controller PublishReleasePre, %+v", err)
	}

	if resp.Code == pbcommon.ErrCode_E_BCS_ALREADY_PUBLISHED {
		return pbcommon.ErrCode_E_OK, ""
	}
	return resp.Code, resp.Message
}

func (act *PublishAction) publishBCSMode(releaseID string) (pbcommon.ErrCode, string) {
	r := &pbbcscontroller.PublishReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: releaseID,
		Operator:  act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("bcscontroller.callTimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%s]| request to bcs-controller, %+v", r.Seq, r)

	resp, err := act.bcsControllerCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to bcs-controller PublishRelease, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *PublishAction) publishGSEPluginMode(releaseID string) (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.PublishReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: releaseID,
		Operator:  act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("gsecontroller.callTimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%s]| request to gse-controller, %+v", r.Seq, r)

	resp, err := act.gseControllerCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to gse-controller PublishRelease, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *PublishAction) publishMultiReleaseData() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.PublishMultiReleaseReq{
		Seq:            act.kit.Rid,
		BizId:          act.req.BizId,
		MultiReleaseId: act.req.MultiReleaseId,
		Operator:       act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.PublishMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager PublishMultiRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	// audit here on release published.
	audit.Audit(int32(pbcommon.SourceType_ST_MULTI_RELEASE), int32(pbcommon.SourceOpType_SOT_PUBLISH),
		act.req.BizId, act.req.MultiReleaseId, act.kit.User, "")

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PublishAction) Do() error {
	// query multi release.
	if errCode, errMsg := act.queryMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.multiRelease.State == int32(pbcommon.ReleaseState_RS_PUBLISHED) {
		// already published.
		return nil
	}
	if act.multiRelease.State != int32(pbcommon.ReleaseState_RS_INIT) {
		return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			"can't publish the multi release which is not init state")
	}

	// query app.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query multi release sub list.
	if errCode, errMsg := act.querySubReleaseList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	// query multi commit sub commit list.
	if errCode, errMsg := act.querySubCommitList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	if len(act.commitIDs) != len(act.releaseIDs) {
		return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			"can't publish the multi release which has inconsonant sub commits and releases")
	}

	// make multi release data published.
	if errCode, errMsg := act.publishMultiReleaseData(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	for _, releaseID := range act.releaseIDs {
		// deploy publish.
		if act.app.DeployType == int32(pbcommon.DeployType_DT_BCS) {
			if errCode, errMsg := act.publishPreBCSMode(releaseID); errCode != pbcommon.ErrCode_E_OK {
				logger.Warnf("PublishMultiRelease[%s]| publish release[%s] pre failed, %+v, %s",
					act.kit.Rid, releaseID, errCode, errMsg)
				continue
			}
			if errCode, errMsg := act.publishBCSMode(releaseID); errCode != pbcommon.ErrCode_E_OK {
				logger.Warnf("PublishMultiRelease[%s]| publish release[%s] failed, %+v, %s",
					act.kit.Rid, releaseID, errCode, errMsg)
				continue
			}
		} else if act.app.DeployType == int32(pbcommon.DeployType_DT_GSE_PLUGIN) ||
			act.app.DeployType == int32(pbcommon.DeployType_DT_GSE) {
			if errCode, errMsg := act.publishPreGSEPluginMode(releaseID); errCode != pbcommon.ErrCode_E_OK {
				logger.Warnf("PublishMultiRelease[%s]| publish release[%s] pre failed, %+v, %s",
					act.kit.Rid, releaseID, errCode, errMsg)
				continue
			}
			if errCode, errMsg := act.publishGSEPluginMode(releaseID); errCode != pbcommon.ErrCode_E_OK {
				logger.Warnf("PublishMultiRelease[%s]| publish release[%s] failed, %+v, %s",
					act.kit.Rid, releaseID, errCode, errMsg)
				continue
			}
		} else {
			return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, "unknow deploy type")
		}
	}

	return nil
}