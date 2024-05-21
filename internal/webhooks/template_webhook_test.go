package webhooks

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

var _ = Describe("Template webhook", func() {
	BeforeEach(func() {
		k8sClient.DeleteAllOf(context.Background(), &cosmov1alpha1.Template{})
		k8sClient.DeleteAllOf(context.Background(), &cosmov1alpha1.ClusterTemplate{})
	})
	Context("when creating Template and the same name Template exist", func() {
		It("should deny", func() {
			ctx := context.Background()
			tmpl1 := cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl1",
				},
			}
			err := k8sClient.Create(ctx, &tmpl1)
			Expect(err).ShouldNot(HaveOccurred())

			tmpl2 := cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl1",
				},
			}
			err = k8sClient.Create(ctx, &tmpl2)
			Expect(err).Should(HaveOccurred())
		})
	})
	Context("when creating ClusterTemplate and the same name Template exist", func() {
		It("should deny", func() {
			ctx := context.Background()
			tmpl1 := cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl1",
				},
			}
			err := k8sClient.Create(ctx, &tmpl1)
			Expect(err).ShouldNot(HaveOccurred())

			tmpl2 := cosmov1alpha1.ClusterTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl1",
				},
			}
			err = k8sClient.Create(ctx, &tmpl2)
			Expect(err).Should(HaveOccurred())
		})
	})
	Context("when creating Template and the same name ClusterTemplate exist", func() {
		It("should deny", func() {
			ctx := context.Background()
			tmpl1 := cosmov1alpha1.ClusterTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl1",
				},
			}
			err := k8sClient.Create(ctx, &tmpl1)
			Expect(err).ShouldNot(HaveOccurred())

			tmpl2 := cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl1",
				},
			}
			err = k8sClient.Create(ctx, &tmpl2)
			Expect(err).Should(HaveOccurred())
		})
	})
})
