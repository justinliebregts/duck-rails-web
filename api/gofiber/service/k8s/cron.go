package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CronJobService defines Kubernetes CronJob management operations.
type CronJobService interface {
	ListCronJobs(ctx context.Context, namespace string) ([]CronJobSummary, error)
	ListExecutedJobs(ctx context.Context, namespace string) ([]JobSummary, error)
	RetrieveJobLogs(ctx context.Context, namespace string, jobName string) ([]string, error)
	TriggerCronJob(ctx context.Context, namespace string, cronjobName string) error
}

type cronJobServiceImpl struct {
	clientset kubernetes.Interface
}

// NewCronJobService returns a new instance of CronJobService.
func NewCronJobService(clientset kubernetes.Interface) CronJobService {
	return &cronJobServiceImpl{
		clientset: clientset,
	}
}

/* k8s permission REQUIRED!

apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: raidcraft
  name: cronjob-executor-role
rules:
  # 1. 允許查看與讀取 CronJob (解決 is forbidden: cannot list)
  - apiGroups: ["batch"]
    resources: ["cronjobs"]
    verbs: ["get", "list", "watch"]

  # 2. 允許手動觸發執行 CronJob (透過 create job)
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["create", "get", "list"]

  - apiGroups: [""] # "" indicates the core API group
    resources: ["pods", "pods/log"]
    verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: bind-cronjob-executor
  namespace: raidcraft
subjects:
  - kind: ServiceAccount
    name: default
    namespace: raidcraft
roleRef:
  kind: Role
  name: cronjob-executor-role
  apiGroup: rbac.authorization.k8s.io


*/

type JobSummary struct {
	Name      string     `json:"name"`
	Status    string     `json:"status"` // Succeeded, Failed, Active
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
}

func (s *cronJobServiceImpl) ListExecutedJobs(ctx context.Context, namespace string) ([]JobSummary, error) {

	jobList, err := s.clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var results []JobSummary
	for _, j := range jobList.Items {
		status := "Active"
		if j.Status.Succeeded > 0 {
			status = "Succeeded"
		} else if j.Status.Failed > 0 {
			status = "Failed"
		}
		var startTime, endTime *time.Time
		if j.Status.StartTime != nil {
			t := j.Status.StartTime.Time
			startTime = &t
		}
		if j.Status.CompletionTime != nil {
			t := j.Status.CompletionTime.Time
			endTime = &t
		}
		results = append(results, JobSummary{
			Name:      j.Name,
			Status:    status,
			StartTime: startTime,
			EndTime:   endTime,
		})
	}
	return results, nil
}

func (s *cronJobServiceImpl) RetrieveJobLogs(ctx context.Context, namespace string, jobName string) ([]string, error) {

	// 1. Find Pods belonging specifically to this Job via label selector
	labelSelector := fmt.Sprintf("job-name=%s", jobName)
	podList, err := s.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods for job %s: %w", jobName, err)
	}

	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("no pods found for job %s in namespace %s", jobName, namespace)
	}

	var logOutputs []string

	// 2. Iterate through each execution Pod and read its container logs
	for _, pod := range podList.Items {
		req := s.clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
		podLogs, err := req.Stream(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to open log stream for pod %s: %w", pod.Name, err)
		}

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		_ = podLogs.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to copy log stream for pod %s: %w", pod.Name, err)
		}

		logOutputs = append(logOutputs, buf.String())
	}

	return logOutputs, nil
}

type CronJobSummary struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Suspend  bool   `json:"suspend"`
}

func (s *cronJobServiceImpl) ListCronJobs(ctx context.Context, namespace string) ([]CronJobSummary, error) {

	cronJobList, err := s.clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []CronJobSummary
	for _, cj := range cronJobList.Items {
		isSuspended := false
		if cj.Spec.Suspend != nil {
			isSuspended = *cj.Spec.Suspend
		}

		results = append(results, CronJobSummary{
			Name:     cj.Name,
			Schedule: cj.Spec.Schedule,
			Suspend:  isSuspended,
		})
	}

	return results, nil
}

func (s *cronJobServiceImpl) TriggerCronJob(ctx context.Context, namespace string, cronjobName string) error {

	// 1. Retrieve existing CronJob to obtain its JobTemplate
	cronJob, err := s.clientset.BatchV1().CronJobs(namespace).Get(ctx, cronjobName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cronjob %s: %w", cronjobName, err)
	}

	// 2. Generate unique Job Name (e.g. cronjobname-manual-timestamp)
	jobName := fmt.Sprintf("%s-manual-%d", cronjobName, time.Now().Unix())
	if len(jobName) > 63 {
		jobName = jobName[:63] // K8s name length limit
	}

	// 3. Construct batchv1.Job from cronJob.Spec.JobTemplate
	manualJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Annotations: map[string]string{
				"cronjob.kubernetes.io/instantiate": "manual",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "batch/v1",
					Kind:       "CronJob",
					Name:       cronJob.Name,
					UID:        cronJob.UID,
				},
			},
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}
	// 4. Create the Job
	_, err = s.clientset.BatchV1().Jobs(namespace).Create(ctx, manualJob, metav1.CreateOptions{})
	return err
}
