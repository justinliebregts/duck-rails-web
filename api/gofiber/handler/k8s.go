package handler

import (
	myMDL "gofiber/model"
	myK8sSRV "gofiber/service/k8s"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type K8sHandler interface {
	TriggerCronJobHandler(c *fiber.Ctx) error
	ListCronJobsHandler(c *fiber.Ctx) error
	ListExecutedJobsHandler(c *fiber.Ctx) error
	RetrieveJobLogHandler(c *fiber.Ctx) error
}

type k8sHandlerImpl struct {
	cronJobSvc myK8sSRV.CronJobService
}

func NewK8sHandler(cronJobSvc myK8sSRV.CronJobService) K8sHandler {
	return &k8sHandlerImpl{
		cronJobSvc: cronJobSvc,
	}
}

type ListCronReq struct {
	Namespace string `params:"ns"`
}

func (r *ListCronReq) Normalize() {
	if r.Namespace == "" {
		r.Namespace = "default"
	}
}

// ListCronJobsHandler godoc
// @Summary List CronJobs
// @Description List CronJobs
// @Tags K8s, admin
// @Accept json
// @Produce json
// @Param ns path string false "Namespace name" default(default)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /info/admin/namespaces/{ns}/cronjobs [get]
func (h *k8sHandlerImpl) ListCronJobsHandler(c *fiber.Ctx) (err error) {
	var resp myMDL.ResponseGeneral

	var req ListCronReq
	if err = c.ParamsParser(&req); err != nil {
		resp = myMDL.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	req.Normalize()

	var cronJobs []myK8sSRV.CronJobSummary
	if cronJobs, err = h.cronJobSvc.ListCronJobs(c.Context(), req.Namespace); err != nil {
		resp = myMDL.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = myMDL.Default200Resp.Copy()
	resp.Data = fiber.Map{
		"cronjobs": cronJobs,
	}
	return c.JSON(resp)
}

type TriggerCronJobReq struct {
	ListCronReq
	CronJob string `params:"cj" validate:"required"`
}

func (r *TriggerCronJobReq) Normalize() {
	r.ListCronReq.Normalize()
}

func (r *TriggerCronJobReq) Validate() (errMsgs []string) {
	if err := validate.Struct(r); err != nil {
		errMsgs = append(errMsgs, err.Error())
	}
	return
}

// TriggerCronJobHandler godoc
// @Summary Trigger CronJob
// @Description Trigger CronJob
// @Tags K8s, admin
// @Accept json
// @Produce json
// @Param ns path string false "Namespace name" default(default)
// @Param cj path string true "CronJob name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /info/admin/namespaces/{ns}/cronjobs/{cj}/trigger [post]
func (h *k8sHandlerImpl) TriggerCronJobHandler(c *fiber.Ctx) (err error) {
	var resp myMDL.ResponseGeneral

	var req TriggerCronJobReq
	if err = c.ParamsParser(&req); err != nil {
		resp = myMDL.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	req.Normalize()

	if errMsgs := req.Validate(); len(errMsgs) > 0 {
		resp = myMDL.Default400Resp.Copy()
		resp.SetErrorAndMessage(strings.Join(errMsgs, ", "))
		return c.Status(resp.Status).JSON(resp)
	}

	if err = h.cronJobSvc.TriggerCronJob(c.Context(), req.Namespace, req.CronJob); err != nil {
		resp = myMDL.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = myMDL.Default200Resp.Copy()
	resp.Data = fiber.Map{
		"result": "CronJob triggered successfully",
	}
	return c.JSON(resp)
}

type ListExecutedJobsReq struct {
	ListCronReq
}

func (r *ListExecutedJobsReq) Normalize() {
	r.ListCronReq.Normalize()
}

// ListExecutedJobsHandler godoc
// @Summary List Executed Jobs
// @Description List Executed Jobs
// @Tags K8s, admin
// @Accept json
// @Produce json
// @Param ns path string false "Namespace name" default(default)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /info/admin/namespaces/{ns}/jobs [get]
func (h *k8sHandlerImpl) ListExecutedJobsHandler(c *fiber.Ctx) (err error) {
	var resp myMDL.ResponseGeneral

	var req ListExecutedJobsReq
	if err = c.ParamsParser(&req); err != nil {
		resp = myMDL.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	req.Normalize()

	var jobs []myK8sSRV.JobSummary
	if jobs, err = h.cronJobSvc.ListExecutedJobs(c.Context(), req.Namespace); err != nil {
		resp = myMDL.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = myMDL.Default200Resp.Copy()
	resp.Data = fiber.Map{
		"jobs": jobs,
	}
	return c.JSON(resp)
}

type ListJobLogsReq struct {
	ListCronReq
	JobName string `params:"jb" validate:"required"`
}

func (r *ListJobLogsReq) Normalize() {
	r.ListCronReq.Normalize()
}

// RetrieveJobLogHandler godoc
// @Summary Retrieve Job Log
// @Description Retrieve Job Log
// @Tags K8s, admin
// @Accept json
// @Produce json
// @Param ns path string false "Namespace name" default(default)
// @Param jb path string true "Job name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /info/admin/namespaces/{ns}/jobs/{jb}/logs [get]
func (h *k8sHandlerImpl) RetrieveJobLogHandler(c *fiber.Ctx) (err error) {
	var resp myMDL.ResponseGeneral

	var req ListJobLogsReq
	if err = c.ParamsParser(&req); err != nil {
		resp = myMDL.Default400Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}
	req.Normalize()

	var jobLogs []string
	if jobLogs, err = h.cronJobSvc.RetrieveJobLogs(c.Context(), req.Namespace, req.JobName); err != nil {
		resp = myMDL.Default500Resp.Copy()
		resp.SetErrorAndMessage(err.Error())
		return c.Status(resp.Status).JSON(resp)
	}

	resp = myMDL.Default200Resp.Copy()
	resp.Data = fiber.Map{
		"namespace": req.Namespace,
		"job":       req.JobName,
		"logs":      jobLogs,
	}
	return c.JSON(resp)
}
