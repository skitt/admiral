/*
SPDX-License-Identifier: Apache-2.0

Copyright Contributors to the Submariner project.

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

package resource_test

import (
	"net"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/submariner-io/admiral/pkg/resource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

var _ = Describe("IsNotFoundErr", func() {
	When("the error is NotFound", func() {
		It("should return true", func() {
			Expect(resource.IsNotFoundErr(apierrors.NewNotFound(schema.GroupResource{}, "foo"))).To(BeTrue())
		})
	})

	When("the error is NoMatchError", func() {
		It("should return true", func() {
			Expect(resource.IsNotFoundErr(&meta.NoKindMatchError{})).To(BeTrue())
		})
	})

	When("the error is ErrGroupDiscoveryFailed", func() {
		Context("and the underlying error is NotFound", func() {
			It("should return true", func() {
				Expect(resource.IsNotFoundErr(&discovery.ErrGroupDiscoveryFailed{
					Groups: map[schema.GroupVersion]error{{}: apierrors.NewNotFound(schema.GroupResource{}, "foo")},
				})).To(BeTrue())
			})
		})

		Context("and the underlying error does not indicate not found", func() {
			It("should return false", func() {
				Expect(resource.IsNotFoundErr(&discovery.ErrGroupDiscoveryFailed{
					Groups: map[schema.GroupVersion]error{{}: apierrors.NewServiceUnavailable("")},
				})).To(BeFalse())
			})
		})
	})

	When("the error does not indicate not found", func() {
		It("should return false", func() {
			Expect(resource.IsNotFoundErr(apierrors.NewForbidden(schema.GroupResource{}, "foo", errors.New("")))).To(BeFalse())
		})
	})
})

var _ = Describe("IsMissingNamespaceErr and ExtractMissingNamespaceFromErr", func() {
	When("the error isn't NotFound", func() {
		It("should return false", func() {
			Expect(resource.IsMissingNamespaceErr(apierrors.NewBadRequest(""))).To(BeFalse())
		})
	})

	When("the error details specify a namespace", func() {
		It("should return true, and the name should be retrievable", func() {
			err := apierrors.NewNotFound(schema.GroupResource{
				Resource: "namespaces",
			}, "missing-ns")
			Expect(resource.IsMissingNamespaceErr(err)).To(BeTrue())
			Expect(resource.ExtractMissingNamespaceFromErr(err)).To(Equal("missing-ns"))
		})
	})

	When("the error details does not specify a namespace", func() {
		It("should return false", func() {
			Expect(resource.IsMissingNamespaceErr(apierrors.NewNotFound(schema.GroupResource{
				Resource: "pods",
			}, "missing"))).To(BeFalse())
		})
	})
})

var _ = Describe("IsTransientErr", func() {
	When("the error is ServerTimeout", func() {
		It("should return true", func() {
			Expect(resource.IsTransientErr(apierrors.NewServerTimeout(schema.GroupResource{}, "get", 5))).To(BeTrue())
		})
	})

	When("the error is TooManyRequests", func() {
		It("should return true", func() {
			Expect(resource.IsTransientErr(apierrors.NewTooManyRequests("", 5))).To(BeTrue())
		})
	})

	When("the error is a wrapped network operation failure", func() {
		It("should return true", func() {
			Expect(resource.IsTransientErr(errors.Wrap(&url.Error{
				Op:  "Get",
				URL: "https://192.168.67.2:8443/api/v1/namespaces",
				Err: &net.OpError{
					Op:  "dial",
					Net: "tcp",
					Addr: &net.TCPAddr{
						IP:   net.IP{192, 168, 67, 2},
						Port: 8443,
					},
					Err: &net.DNSError{},
				},
			}, "wrapped"))).To(BeTrue())
		})
	})

	When("the error is not transient", func() {
		It("should return false", func() {
			Expect(resource.IsTransientErr(apierrors.NewBadRequest(""))).To(BeFalse())
		})
	})

	When("the error is nil", func() {
		It("should return false", func() {
			Expect(resource.IsTransientErr(nil)).To(BeFalse())
		})
	})
})
