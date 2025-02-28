// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db"
	sql "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	IstioNamespace       = "istio-system"
	CertManagerNamesapce = "cert-manager"
	OpenEbsNamespace     = "openebs"
	SecretNamespace      = "secrets"
	DockerRegistrySecret = "docker-registry-secret"
)

// cert manager config for the gateway
const (
	IstioRootClusterIssuer            = "istio-self-sign-issuer"
	IstioRootClusterIssuerCertificate = "istio-ca"
	IstioK8SCSR                       = "istio-ca-issuer"
	CLUSTER_ISSUER                    = "ClusterIssuer"   // apiVersion Kind
	CLUSTER_ISSUER_API_VERSION        = "cert-manager.io" // api version for the cert manager
	CLUSTER_ROOT_SECRET               = "root-secret"     // used for the gateway configuration
)

const (
	IngressIKSNodeLabelKey = "nodegroupName"
)

// Task represents a unit of work to be executed.
type Task struct {
	ID                string
	Name              string
	Description       string
	Func              func(input *TaskRunContext) (interface{}, error)
	SubDeployment     *Deployment
	Input             interface{}
	Output            interface{}
	Status            TaskStatus
	ErrorMessage      string
	Dependent         []*Task
	DeploymentContext Deployment
	Context           TaskContext
}

type TaskContext struct {
	TimeOut     time.Duration
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	DoneCh      chan struct{}
}

type TaskRunContext struct {
	Task              Task
	Input             interface{}
	DeploymentContext Deployment
	SqlConn           *pgxpool.Conn
	SqlModel          *sql.Queries
	K8sClient         k8s.K8sClient
}

// TaskStatus represents the status of a task.
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskWaitingForDependency
	TaskInProgress
	TaskCompleted
	TaskFailed
)

// NewTask creates a new task with the given name and function.
func NewTask(name string, f func(input *TaskRunContext) (interface{}, error), input interface{}, dependents []*Task) *Task {
	return &Task{
		ID:        uuid.New().String(),
		Name:      name,
		Func:      f,
		Status:    TaskPending,
		Input:     input,
		Dependent: dependents, //[]*Task{},
		Context: TaskContext{
			CreatedAt: time.Now(),
			DoneCh:    make(chan struct{}),
		},
	}
}

// func NewTaskWithSubDeployment(name string, subDeployment *Deployment, input interface{}, dependents []*Task) *Task {
// 	return &Task{
// 		ID:            uuid.New().String(),
// 		Name:          name,
// 		SubDeployment: subDeployment,
// 		Status:        TaskPending,
// 		Input:         input,
// 		Dependent:     dependents, //[]*Task{},
// 		Context: TaskContext{
// 			CreatedAt: time.Now(),
// 			doneCh:    make(chan struct{}),
// 		},
// 	}
// }

// Execute executes the task's function and updates its status.
func (t *Task) Execute() {
	if t.Status != TaskPending {
		// Task has already been executed or is in progress
		return
	}

	pool := t.DeploymentContext.Context.SqlPool
	if err := pool.Ping(context.Background()); err != nil {
		fmt.Printf("Error: %+v", err)
		LogTaskFailure(t, nil, err)
		return
	}
	conn, err := pool.Acquire(context.Background())

	defer conn.Release()
	if err != nil {
		LogTaskFailure(t, nil, fmt.Errorf("failed to initialize database: %+v", err))
		return
	}
	model := sql.New(pool)

	// initialize gRPC client to IKS
	var k8sClient k8s.K8sClient
	k8sClient.ClusterID = t.DeploymentContext.Context.IksClusterId
	log.Printf("Config : %+v", t.DeploymentContext.Context.Conf)
	log.Printf("Cluster Id: %+v ", k8sClient.ClusterID)
	err = k8sClient.GetIksClient(t.DeploymentContext.Context.Conf)
	if err != nil {
		log.Printf("%s: Task %s from Deployment: %s failed: unable to get the K8s ClientSet %+v\n", t.ID, t.Name, t.DeploymentContext.Name, err)
		LogTaskFailure(t, model, err)
		return
	}
	defer k8sClient.GrpcClientConn.Close()

	log.Printf("Executing Task: %s from deployment: %s\n", t.Name, t.DeploymentContext.Name)
	_, _ = model.CreateDeploymentTask(context.TODO(), sql.CreateDeploymentTaskParams{
		ID:           t.ID,
		DeploymentID: t.DeploymentContext.ID,
		Name:         t.Name,
	})

	t.Status = TaskInProgress
	t.Context.StartedAt = time.Now()

	model.UpdateDeploymentTaskStatusAsWaitingForUpstream(context.Background(), sql.UpdateDeploymentTaskStatusAsWaitingForUpstreamParams{
		ID:                t.ID,
		StatusDisplayName: pgtype.Text{String: "Waiting for UpStream", Valid: true},
		StatusMessage:     pgtype.Text{String: "Waiting for UpStream", Valid: true},
	})

	// Wait for dependencies to complete
	for _, dep := range t.Dependent {
		<-dep.Context.DoneCh
		if dep.Status == TaskFailed {

			errMessage := fmt.Sprintf("Task %s from deployment: %s cannot execute due to failed dependency %s\n", t.Name, t.DeploymentContext.Name, dep.Name)
			log.Println(errMessage)
			t.Status = TaskFailed
			t.DeploymentContext.Status = DeploymentFailed
			t.Context.CompletedAt = time.Now()

			model.UpdateDeploymentTaskStatusAsUpstreamFailed(context.Background(), sql.UpdateDeploymentTaskStatusAsUpstreamFailedParams{
				ID:                t.ID,
				StatusDisplayName: pgtype.Text{String: "Upstream Failed", Valid: true},
				StatusMessage:     pgtype.Text{String: fmt.Sprintf("Dependency Task %s failed", dep.Name), Valid: true},
				ErrorMessage:      pgtype.Text{String: errMessage, Valid: true},
			})
			close(t.Context.DoneCh)
			return
		}
	}

	model.UpdateDeploymentTaskStatusAsRunning(context.Background(), sql.UpdateDeploymentTaskStatusAsRunningParams{
		ID:                t.ID,
		StatusDisplayName: pgtype.Text{String: fmt.Sprintf("%s in Progress", t.Name), Valid: true},
		StatusMessage:     pgtype.Text{String: fmt.Sprintf("Deployment Task %s in Progress", t.Name), Valid: true},
	})

	model.UpdateDeploymentStatusAsRunning(context.Background(), sql.UpdateDeploymentStatusAsRunningParams{
		ID:                t.DeploymentContext.ID,
		StatusDisplayName: pgtype.Text{String: "Deployment in progress", Valid: true},
		StatusMessage:     pgtype.Text{String: fmt.Sprintf("Deployment Task %s in Progress", t.Name), Valid: true},
	})

	var output any
	// Execute the sub-deployment if it is defined
	input := TaskRunContext{
		Task:              *t,
		Input:             t.Input,
		DeploymentContext: t.DeploymentContext,
		K8sClient:         k8sClient,
		SqlModel:          model,
		SqlConn:           conn,
	}

	log.Printf("%s: Task %s from deployment: %s started...\n", t.ID, t.DeploymentContext.Name, t.Name)
	output, err = t.Func(&input)

	if err != nil {
		LogTaskFailure(t, model, err)
		return
	}

	log.Printf("%s: Task %s from deployment: %s completed\n", t.ID, t.DeploymentContext.Name, t.Name)
	t.Status = TaskCompleted
	t.Output = output

	outputByte, err := json.Marshal(output)
	if err != nil {
		log.Println("Not able to covert the output to json bytes")
	}

	model.UpdateDeploymentTaskStatusAsSuccess(context.Background(), sql.UpdateDeploymentTaskStatusAsSuccessParams{
		ID:            t.ID,
		OutputPayload: outputByte,
	})
	t.DeploymentContext.Output = outputByte
	// Signal completion to dependent tasks
	close(t.Context.DoneCh)
}

// Log Task failure
func LogTaskFailure(t *Task, model *sql.Queries, err error) {
	log.Printf("%s: Task %s from Deployment: %s failed: %+v\n", t.ID, t.Name, t.DeploymentContext.Name, err)
	t.Status = TaskFailed
	t.DeploymentContext.Status = DeploymentFailed
	t.DeploymentContext.ErrorMessage = t.DeploymentContext.ErrorMessage + fmt.Sprintf("%s/n", err)
	close(t.Context.DoneCh)
	if model != nil {
		ierr := model.UpdateDeploymentTaskStatusAsFailed(context.Background(), sql.UpdateDeploymentTaskStatusAsFailedParams{
			ID:                t.ID,
			StatusDisplayName: pgtype.Text{String: "Failed", Valid: true},
			StatusMessage:     pgtype.Text{String: fmt.Sprintf("Deployment failed at the task %s", t.Name), Valid: true},
			ErrorMessage:      pgtype.Text{String: fmt.Sprintf("%v", err), Valid: true},
		})
		if ierr != nil {
			log.Printf("%s: Task %s from Deployment: %s failed: %+v\n failed to commit the status to backed database. Error Message: %+v", t.ID, t.Name, t.DeploymentContext.Name, err, ierr)
		}

	}
	log.Printf("%s: Task %s from Deployment: %s failed: %+v", t.ID, t.Name, t.DeploymentContext.Name, err)
}

// Add Input to the tasks
func (t *Task) WithInput(v interface{}) {
	t.Input = v
}

// Get Task output
// func GetTaskOutput(ctx TaskRunContext, targetTaskName string) any {
// 	value, ok := ctx.DeploymentContext.Context.TaskIndex[targetTaskName]

// 	if !ok {
// 		return nil
// 	}
// 	return ctx.DeploymentContext.Tasks[value].Output
// }

func (t *TaskRunContext) GetTaskOutput(targetTaskName string) any {
	value, ok := t.DeploymentContext.Context.TaskIndex[targetTaskName]

	if !ok {
		return nil
	}
	return t.DeploymentContext.Tasks[value].Output
}

func (t *TaskRunContext) GetDeploymentOutput(deploymentId string) ([]byte, error) {

	deployment, err := t.SqlModel.GetDeployment(context.Background(), sql.GetDeploymentParams{
		ID: deploymentId,
	})
	if err != nil {
		log.Printf("Failed to get the output of the deployment with id %s. \nError Message: %+v", deploymentId, err)
		return nil, err
	}
	return deployment.OutputPayload, nil
}

// TaskStatus represents the status of a task.
type DeploymentStatus int

const (
	DeploymentPending DeploymentStatus = iota
	DeploymentInProgress
	DeploymentCompleted
	DeploymentFailed
)

type DeploymentInputContext struct {
	ID                 string
	SqlPool            *pgxpool.Pool
	IsSubDeployment    bool
	ParentDeploymentId string
	Conf               *config.Config
}

type DeploymentContext struct {
	TaskIndex map[string]int
	Sql       *sql.Queries
	SqlPool   *pgxpool.Pool
	// ResourceId      string // Will be used only when deployment runs for the CREATE action. For rest all resources, the resource id will come from the input payload.
	IsSubDeployment    bool
	ParentDeploymentId string
	RandomSuffix       string
	Conf               *config.Config
	IksClusterId       *pb.ClusterID
	WorkspaceId        string
	ServiceId          string
}

// Deployment represents a series of tasks to be executed in order.
type Deployment struct {
	ID           string
	Name         string
	Tasks        []*Task
	RawInput     []byte
	Output       []byte
	Status       DeploymentStatus
	ErrorMessage string
	Context      DeploymentContext
}

// func GetIksClusterID(sqlModel *sql.Queries, cloudAccountID, workspaceID, serviceID string) (*pb.ClusterID, error) {
// 	clusterID := &pb.ClusterID{
// 		CloudAccountId: cloudAccountID,
// 	}

// 	if workspaceID != "" {
// 		workspaceData, err := sqlModel.GetWorkspaceById(context.TODO(), workspaceID)
// 		if err != nil {
// 			return nil, fmt.Errorf("no Match found for the workspace id: %s", workspaceID)
// 		}
// 		clusterID.Clusteruuid = workspaceData.IksID
// 	} else if serviceID != "" {
// 		workspaceData, err := sqlModel.GetClusterIdFromServiceId(context.TODO(), pgtype.Text{String: serviceID, Valid: true})
// 		if err != nil {
// 			return nil, fmt.Errorf("no Match found for the workspace id: %s", serviceID)
// 		}
// 		clusterID.Clusteruuid = workspaceData.IksID
// 	}

// 	return clusterID, nil
// }

// NewDeployment creates a new deployment with the given name.
func NewDeployment(ctx DeploymentInputContext, name string) (*Deployment, error) {

	if err := ctx.SqlPool.Ping(context.Background()); err != nil {
		fmt.Printf("Error: %+v", err)
	} else {
		log.Printf("Ping success: %s \n", name)
	}

	// Connect to DB and get the user inputs
	// conn, err := db.AcquireConn(ctx.SqlPool)
	// if err != nil {
	// 	log.Fatal("Error connecting to the database...", err)
	// }
	// defer conn.Release()
	sqlModel := sql.New(ctx.SqlPool)

	data, err := sqlModel.UpdateDeploymentStatusAsRunning(context.TODO(), sql.UpdateDeploymentStatusAsRunningParams{
		ID:                ctx.ID,
		StatusDisplayName: pgtype.Text{String: "Deployment in progress", Valid: true},
		StatusMessage:     pgtype.Text{String: fmt.Sprintf("Deployment: %s in progress", name), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("noEntryFound for the deployment id: %s", ctx.ID)
	}
	log.Printf("NewDeployment data: %+v", data)

	clusterID := &pb.ClusterID{}
	if data.ChangeIndicator != pb.DpaiDeploymentChangeIndicator_DPAI_CREATE.String() || (data.WorkspaceID.String != "" && data.ServiceID.String != "") {
		clusterID, err = k8s.GetIksClusterID(sqlModel, data.WorkspaceID.String, data.ServiceID.String)
		if err != nil {
			return nil, err
		}
	}
	log.Printf("Context Configuration: %+v", ctx.Conf)
	log.Printf("Inside New Deployment clusterID: %+v", clusterID)

	return &Deployment{
		ID:       data.ID,
		Name:     name,
		Tasks:    []*Task{},
		RawInput: data.InputPayload,
		Context: DeploymentContext{
			SqlPool:            ctx.SqlPool,
			TaskIndex:          map[string]int{},
			IsSubDeployment:    ctx.IsSubDeployment,
			ParentDeploymentId: ctx.ParentDeploymentId,
			RandomSuffix:       data.ID[len(data.ID)-20:],
			Conf:               ctx.Conf,
			IksClusterId:       clusterID,
			WorkspaceId:        data.WorkspaceID.String,
			ServiceId:          data.ServiceID.String,
		},
	}, nil
}

// AddTask adds a task to the deployment.
func (d *Deployment) AddTask(task *Task) {

	for _, existingTask := range d.Tasks {
		if existingTask.Name == task.Name {
			log.Printf("task Name '%s' already exists, skipping...\n", task.Name)
		}
	}
	d.Tasks = append(d.Tasks, task)
	// task.DeploymentContext = *d
	d.Context.TaskIndex[task.Name] = len(d.Tasks) - 1
}

// AddTask adds a task to the deployment.
func (d *Deployment) AddTasks(tasks []*Task) {
	for _, task := range tasks {
		d.AddTask(task)
	}
}

// Get the input same as provided by the user
func (d *Deployment) GetDeploymentInput(value any) (any, error) {
	err := json.Unmarshal(d.RawInput, value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Run executes the deployment and runs all tasks in order.
func (d *Deployment) Run() ([]byte, error) {

	log.Printf("Deployments: %v", d)

	for _, task := range d.Tasks {
		task.DeploymentContext = *d
		log.Printf("Task Deployments: %v", task.DeploymentContext)
		go task.Execute()
	}

	// Wait for the completion of all tasks in the deployment
	for _, task := range d.Tasks {
		<-task.Context.DoneCh
		// Assign the last completing task's output as a output
		// of the deployment.
		if task.Status == TaskCompleted {
			d.Output = task.DeploymentContext.Output
		} else if task.Status == TaskFailed {
			d.Status = DeploymentFailed
			d.ErrorMessage = task.DeploymentContext.ErrorMessage
		}
	}

	conn, err := db.AcquireConn(d.Context.SqlPool)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %+v", err)
	}
	defer conn.Release()
	sqlModel := sql.New(d.Context.SqlPool)

	if d.Status == DeploymentFailed {
		sqlModel.UpdateDeploymentStatusAsFailed(context.Background(), sql.UpdateDeploymentStatusAsFailedParams{
			ID:                d.ID,
			ErrorMessage:      pgtype.Text{String: d.ErrorMessage, Valid: true},
			StatusDisplayName: pgtype.Text{String: "Deployment Failed", Valid: true},
			StatusMessage:     pgtype.Text{String: fmt.Sprintf("deployment failed with the error %v", d.ErrorMessage), Valid: true},
		})
		return d.Output, fmt.Errorf("deployment failed with the error %v", d.ErrorMessage)
	} else {
		sqlModel.UpdateDeploymentStatusAsSuccess(context.Background(), sql.UpdateDeploymentStatusAsSuccessParams{
			ID:                d.ID,
			OutputPayload:     d.Output,
			StatusDisplayName: pgtype.Text{String: "Deployment Completed", Valid: true},
			StatusMessage:     pgtype.Text{String: "Deployment Completed", Valid: true},
		})
	}

	return d.Output, nil
}

type DeploymentCleanUpParams struct {
	// This parameter should be provided only for the workspace deployment
	NodeGroupId   string
	Namespace     string
	IksClusterId  pb.ClusterID
	DeleteCluster bool
}

func (d *Deployment) CleanUp(params *DeploymentCleanUpParams) error {

	log.Printf("Performing tthe cleanup operation for the deployment: %s", d.ID)

	var k8sClient k8s.K8sClient

	if d.Context.IksClusterId.Clusteruuid != "" {
		k8sClient.ClusterID = d.Context.IksClusterId
		err := k8sClient.GetIksClient(d.Context.Conf)
		if err != nil {
			log.Printf("Cleanup Task from Deployment: %s failed: unable to get the K8s ClientSet %+v\n", d.Name, err)
			return err
		}
	}
	defer k8sClient.GrpcClientConn.Close()

	if params.IksClusterId.CloudAccountId != "" {
		// call delete IKS cluster api.
		log.Println("Call IKS api to delete the IKS cluster.")
		return nil
	}

	var nodeGroupError error
	if params.NodeGroupId != "" && d.Context.ParentDeploymentId == "" {
		// TODO: Code to cleanup the nodegroup.
		log.Println("Cleaned up the nodegroup")
	}

	var namespaceError error
	if params.Namespace != "" && d.Context.ParentDeploymentId == "" {
		_, namespaceError = k8sClient.DeleteNamespace(params.Namespace, true)
		log.Println("Cleaned up the namespace")
	}

	secretError := k8sClient.DeleteSecret("secrets", d.ID, true)

	if nodeGroupError != nil || namespaceError != nil || secretError != nil {
		return fmt.Errorf("cleanupError: Error while cleaning up the failed deployment. Manual action needed")
	}

	return nil
}
