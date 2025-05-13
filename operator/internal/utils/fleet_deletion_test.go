package utils

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type FakeFleetDeleteChecker struct {
	DeletionState map[string]bool
}

func (f FakeFleetDeleteChecker) isDeleteAllowed(ctx context.Context, server *networkv1alpha1.Server, c *client.Client) (bool, error) {
	return f.DeletionState[server.Name], nil
}

var _ = Describe("Fleet Utility Testing", func() {
	Context("When finding the server to delete", func() {
		ctx := context.Background()
		It("Find oldest", func() {
			By("Setup objects")
			baseTime := time.Now()
			fake := FakeFleetDeleteChecker{DeletionState: make(map[string]bool)}
			servers := networkv1alpha1.ServerList{Items: []networkv1alpha1.Server{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server1",
						CreationTimestamp: metav1.Time{Time: baseTime},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server2",
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Hour)},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server3",
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Minute)},
					},
				},
			}}
			By("Find the oldest server")
			server, err := getOldestServer(ctx, &servers, false, nil, fake)
			Expect(err).ToNot(HaveOccurred())
			Expect(servers.Items).To(HaveLen(3))
			Expect(server.Name).To(Equal("server1"))
		})

		It("Find oldest with delete allowed", func() {
			By("Setup objects")
			baseTime := time.Now()
			fake := FakeFleetDeleteChecker{DeletionState: make(map[string]bool)}
			servers := networkv1alpha1.ServerList{Items: []networkv1alpha1.Server{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server1",
						CreationTimestamp: metav1.Time{Time: baseTime}, // 0
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server2",
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Hour)}, // 2
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server3",
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Minute)}, // 1
					},
				},
			}}
			By("Find oldest deletable server")
			fake.DeletionState["server2"] = true
			fake.DeletionState["server3"] = true
			server, err := getOldestServer(ctx, &servers, true, nil, fake)
			Expect(err).ToNot(HaveOccurred())
			Expect(servers.Items).To(HaveLen(3))
			Expect(server.Name).To(Equal("server3"))
		})

		It("Find youngest", func() {
			By("Setup objects")
			baseTime := time.Now()
			fake := FakeFleetDeleteChecker{DeletionState: make(map[string]bool)}
			servers := networkv1alpha1.ServerList{Items: []networkv1alpha1.Server{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server1",
						CreationTimestamp: metav1.Time{Time: baseTime}, // 0
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server2",
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Hour)}, // 2
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server3",
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Minute)}, // 1
					},
				},
			}}
			By("Find the oldest server")
			server, err := getNewestServer(ctx, &servers, false, nil, fake)
			Expect(err).ToNot(HaveOccurred())
			Expect(servers.Items).To(HaveLen(3))
			Expect(server.Name).To(Equal("server2"))
		})

		It("Find oldest with delete allowed", func() {
			By("Setup objects")
			baseTime := time.Now()
			fake := FakeFleetDeleteChecker{DeletionState: make(map[string]bool)}
			servers := networkv1alpha1.ServerList{Items: []networkv1alpha1.Server{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server1",
						CreationTimestamp: metav1.Time{Time: baseTime},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server2",
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Hour)},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "server3",
						CreationTimestamp: metav1.Time{Time: baseTime.Add(time.Minute)},
					},
				},
			}}
			By("Find youngest deletable server")
			fake.DeletionState["server2"] = true
			fake.DeletionState["server3"] = true
			server, err := getOldestServer(ctx, &servers, true, nil, fake)
			Expect(err).ToNot(HaveOccurred())
			Expect(servers.Items).To(HaveLen(3))
			Expect(server.Name).To(Equal("server3"))
		})
	})
})
