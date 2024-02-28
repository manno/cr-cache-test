package cachetest

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

var (
	cl client.Client
)

var _ = Describe("Cache", func() {
	BeforeEach(func() {
		systemNamespace := "default"
		cluster, err := cluster.New(cfg, func(clusterOptions *cluster.Options) {
			clusterOptions.Scheme = scheme.Scheme
			clusterOptions.Cache = cache.Options{
				ByObject: map[client.Object]cache.ByObject{
					&corev1.ConfigMap{}: {
						Namespaces: map[string]cache.Config{
							systemNamespace: {},
						},
					},
				},
				DefaultNamespaces: map[string]cache.Config{cache.AllNamespaces: {}},
			}
		})
		Expect(err).NotTo(HaveOccurred())

		go func() {
			err := cluster.GetCache().Start(ctx)
			Expect(err).NotTo(HaveOccurred())
		}()
		cluster.GetCache().WaitForCacheSync(ctx)

		cl = cluster.GetClient()
	})

	It("respect the cache config", func() {
		By("list and get work for objects in the namespace from ByObject", func() {
			err := cl.Create(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}})
			Expect(err).NotTo(HaveOccurred())

			err = cl.Get(ctx, client.ObjectKey{Name: "test", Namespace: "default"}, &corev1.ConfigMap{})
			Expect(err).NotTo(HaveOccurred())

			err = cl.List(ctx, &corev1.ConfigMapList{}, client.InNamespace("default"))
			Expect(err).NotTo(HaveOccurred())
		})

		By("list and get work for all namespaces, if objects are not in ByObject", func() {
			err := cl.Create(ctx, &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}})
			Expect(err).NotTo(HaveOccurred())

			err = cl.Get(ctx, client.ObjectKey{Name: "test", Namespace: "default"}, &corev1.ServiceAccount{})
			Expect(err).NotTo(HaveOccurred(), "get failed")

			err = cl.List(ctx, &corev1.ServiceAccountList{}, client.InNamespace("default"))
			Expect(err).NotTo(HaveOccurred(), "list failed")
		})

	})
})
