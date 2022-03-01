package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestAddPodLogger(t *testing.T) {
	testPod := testPod()

	addPodLogger(testPod)

	assert.Equal(t, 1, podLogger.PodSize)

	for _, v := range podLogger.PodLoggerTemplate {
		assert.Equal(t, 1, v.ContainerSize)
		assert.Equal(t, "test-pod", v.PodName)
		assert.Equal(t, "nginx", v.PodLoggerContainer[0].Name)
		assert.Equal(t, "nginx", v.PodLoggerContainer[0].Image)
	}
}

func TestUpdatePodLogger(t *testing.T) {
	testPod := testPod()

	addPodLogger(testPod)
	testPod.Spec.Containers[0].Image = "nginx:alpine"
	updatePodLogger(testPod)

	assert.Equal(t, "nginx:alpine", podLogger.PodLoggerTemplate[0].PodLoggerContainer[0].Image)
}

func testPod() *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "nginx",
					Image:           "nginx",
					ImagePullPolicy: "Always",
				},
			},
		},
	}
}
