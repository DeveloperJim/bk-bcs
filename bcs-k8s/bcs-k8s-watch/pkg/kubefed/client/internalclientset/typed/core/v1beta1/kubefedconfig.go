/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "bk-bcs/bcs-k8s/bcs-k8s-watch/pkg/kubefed/apis/core/v1beta1"
	scheme "bk-bcs/bcs-k8s/bcs-k8s-watch/pkg/kubefed/client/internalclientset/scheme"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KubeFedConfigsGetter has a method to return a KubeFedConfigInterface.
// A group's client should implement this interface.
type KubeFedConfigsGetter interface {
	KubeFedConfigs(namespace string) KubeFedConfigInterface
}

// KubeFedConfigInterface has methods to work with KubeFedConfig resources.
type KubeFedConfigInterface interface {
	Create(*v1beta1.KubeFedConfig) (*v1beta1.KubeFedConfig, error)
	Update(*v1beta1.KubeFedConfig) (*v1beta1.KubeFedConfig, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.KubeFedConfig, error)
	List(opts v1.ListOptions) (*v1beta1.KubeFedConfigList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.KubeFedConfig, err error)
	KubeFedConfigExpansion
}

// kubeFedConfigs implements KubeFedConfigInterface
type kubeFedConfigs struct {
	client rest.Interface
	ns     string
}

// newKubeFedConfigs returns a KubeFedConfigs
func newKubeFedConfigs(c *CoreV1beta1Client, namespace string) *kubeFedConfigs {
	return &kubeFedConfigs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kubeFedConfig, and returns the corresponding kubeFedConfig object, and an error if there is any.
func (c *kubeFedConfigs) Get(name string, options v1.GetOptions) (result *v1beta1.KubeFedConfig, err error) {
	result = &v1beta1.KubeFedConfig{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kubefedconfigs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KubeFedConfigs that match those selectors.
func (c *kubeFedConfigs) List(opts v1.ListOptions) (result *v1beta1.KubeFedConfigList, err error) {
	result = &v1beta1.KubeFedConfigList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kubefedconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kubeFedConfigs.
func (c *kubeFedConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kubefedconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a kubeFedConfig and creates it.  Returns the server's representation of the kubeFedConfig, and an error, if there is any.
func (c *kubeFedConfigs) Create(kubeFedConfig *v1beta1.KubeFedConfig) (result *v1beta1.KubeFedConfig, err error) {
	result = &v1beta1.KubeFedConfig{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kubefedconfigs").
		Body(kubeFedConfig).
		Do().
		Into(result)
	return
}

// Update takes the representation of a kubeFedConfig and updates it. Returns the server's representation of the kubeFedConfig, and an error, if there is any.
func (c *kubeFedConfigs) Update(kubeFedConfig *v1beta1.KubeFedConfig) (result *v1beta1.KubeFedConfig, err error) {
	result = &v1beta1.KubeFedConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kubefedconfigs").
		Name(kubeFedConfig.Name).
		Body(kubeFedConfig).
		Do().
		Into(result)
	return
}

// Delete takes name of the kubeFedConfig and deletes it. Returns an error if one occurs.
func (c *kubeFedConfigs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kubefedconfigs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kubeFedConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kubefedconfigs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched kubeFedConfig.
func (c *kubeFedConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.KubeFedConfig, err error) {
	result = &v1beta1.KubeFedConfig{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kubefedconfigs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
