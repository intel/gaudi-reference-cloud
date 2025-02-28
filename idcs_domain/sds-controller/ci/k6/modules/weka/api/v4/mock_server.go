// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v4

import (
	"net/http"
	"sync"

	v4 "github.com/labstack/echo/v4"
)

type MockWekaAPI struct {
	deleteCalled bool
	lock         sync.Mutex
	ServerInterface
}

var organization interface{} = map[string]interface{}{
	"total_quota":     20000002048,
	"id":              1,
	"ssd_quota":       10000003072,
	"ssd_allocated":   0,
	"name":            "ORG-1",
	"total_allocated": 0,
	"uid":             "uid_string",
}

var user interface{} = map[string]interface{}{
	"uid":      "uid_string",
	"org_id":   0,
	"source":   "Internal",
	"username": "admin",
	"role":     "ClusterAdmin",
}

var bucket interface{} = map[string]interface{}{
	"hard_limit_bytes": 5555,
	"name":             "bucketName",
	"path":             "bucket",
	"used_bytes":       2222,
	"role":             "S3",
}

var bucketPolicy interface{} = map[string]interface{}{
	"policy": "upload",
}

var lifecycleRule interface{} = map[string]interface{}{
	"enabled":     true,
	"expiry_days": "60",
	"id":          "lrId",
	"prefix":      "prefix",
}

var s3Policy interface{} = map[string]interface{}{
	"policy": map[string]interface{}{
		"content": map[string]interface{}{
			"Statement": []interface{}{
				map[string]interface{}{
					"Resource": []string{
						"arn:aws:s3:::*",
					},
					"Sid": "",
					"Action": []string{
						"s3:GetObject",
						"s3:ListBucket",
					},
					"Effect": "Allow",
				},
			},
			"Version": "2012-10-17",
		},
		"name": "policyName",
	},
	"session_id": "sessions",
}

var fsGroup interface{} = map[string]interface{}{
	"name":                 "fs-group1",
	"start_demote":         1800,
	"target_ssd_retention": 86400,
	"uid":                  "uid_string",
	"id":                   "FSGroupId<0>",
}

var filesystem interface{} = map[string]interface{}{
	"id":                     "FSId<2>",
	"auto_max_files":         false,
	"used_ssd_data":          0,
	"name":                   "fs1",
	"uid":                    "uid_string",
	"is_removing":            false,
	"group_id":               "FSGroupId<0>",
	"is_creating":            true,
	"free_total":             1073737728,
	"is_encrypted":           true,
	"metadata_budget":        4202496,
	"used_total_data":        0,
	"used_total":             4096,
	"ssd_budget":             1073741824,
	"is_ready":               false,
	"group_name":             "default",
	"available_total":        1073741824,
	"status":                 "CREATING",
	"used_ssd_metadata":      4096,
	"auth_required":          true,
	"available_ssd_metadata": 4202496,
	"total_budget":           1073741824,
	"used_ssd":               4096,
	"obs_buckets": []interface{}{
		map[string]interface{}{
			"uid":            "237",
			"state":          "ACTIVE",
			"status":         "UP",
			"detachProgress": 0,
			"obsId":          "ObjectStorageId<0>",
			"mode":           "WRITABLE",
			"name":           "OBS_1",
		},
	},
	"available_ssd": 1073741824,
	"free_ssd":      1073737728,
}

func (*MockWekaAPI) GetClusterStatus(ctx v4.Context) error {
	var usableCapacity uint64 = 18446744073709551615
	var clusterData interface{} = map[string]interface{}{
		"overlay": map[string]interface{}{
			"client_nodes_safety_histogram": []int{
				0, 2, 3, 4,
			},
			"branching_factor":           40,
			"clients_branching_factor":   100,
			"client_nodes_at_risk":       0,
			"client_nodes_not_supported": 0,
		},
		"activity": map[string]interface{}{
			"obs_upload_bytes_per_second":   0,
			"sum_bytes_read":                0,
			"num_writes":                    0,
			"obs_download_bytes_per_second": 0,
			"sum_bytes_written":             0,
			"num_reads":                     0,
			"num_ops":                       0,
		},
		"hot_spare":         1,
		"io_status":         "STARTED",
		"status":            "OK",
		"last_init_failure": "",
		"drives": map[string]interface{}{
			"active": 7,
			"total":  7,
		},
		"name":                              "stewie",
		"upgrade":                           "",
		"io_status_changed_time":            "2021-03-10T09:46:02.245149Z",
		"short_drive_grace_on_failure_secs": 10,
		"io_nodes": map[string]interface{}{
			"active": 7,
			"total":  7,
		},
		"cloud": map[string]interface{}{
			"enabled": true,
			"healthy": true,
			"proxy":   "",
		},
		"release_hash": "d79aaa56d95d9beb02b15783dede1ec79d4c273b",
		"rebuild": map[string]interface{}{
			"requiredFDsForRebuild": 4,
			"unavailablePercent":    0,
			"enoughActiveFDs":       true,
			"totalCopiesMiB":        0,
			"unavailableMiB":        0,
			"progressPercent":       0,
			"stripeDisks":           7,
			"numActiveFDs":          7,
			"totalCopiesDoneMiB":    0,
			"protectionState": []interface{}{
				map[string]interface{}{
					"percent":     100,
					"numFailures": 0,
					"MiB":         403200,
				},
			},
		},
		"init_stage":                       "INITIALIZED",
		"failure_domains_enabled":          true,
		"long_drive_grace_on_failure_secs": 360,
		"hosts": map[string]interface{}{
			"total_count": 7,
			"backends": map[string]interface{}{
				"active": 7,
				"total":  7,
			},
			"active_count": 7,
			"clients": map[string]interface{}{
				"active": 0,
				"total":  0,
			},
		},
		"last_init_failure_code": "",
		"stripe_data_drives":     5,
		"release":                "3.11.1.6926-f6e1fbb96d03ef76b64bfd4d8c2c366a",
		"buckets_info": map[string]interface{}{
			"averageFillLevelPPMMinSinceLastEvent": 0,
			"thinProvisionState": map[string]interface{}{
				"usableWritable":  521109504,
				"totalSSDBudgets": 446644224,
				"shrinkageFactor": map[string]interface{}{
					"val": 4096,
				},
			},
			"averageFillLevelPPM":             5269,
			"shrunkAtGeneration":              "ConfigGeneration<INVALID>",
			"placementAllocationThresholdPPM": 450000,
			"maxPrefetchRPCs":                 256,
		},
		"active_alerts_count": 0,
		"time": map[string]interface{}{
			"cluster_time":                     "2021-03-10T09:52:14.5068161Z",
			"allowed_clock_skew_secs":          60,
			"cluster_local_utc_offset_seconds": 1800,
			"cluster_local_utc_offset":         "+03:00",
		},
		"net": map[string]interface{}{
			"link_layer": "ETH",
		},
		"buckets": map[string]interface{}{
			"active":                  105,
			"global_flush_generation": 1,
			"global_flush_status":     "NONE",
			"total":                   105,
			"flush_finished":          105,
			"shutdown_finished":       105,
		},
		"scrubber_bytes_per_sec":  134217728,
		"init_stage_changed_time": "2021-03-10T09:45:42.599016Z",
		"capacity": map[string]interface{}{
			"total_bytes":         1829454741504,
			"hot_spare_bytes":     305009786880,
			"unprovisioned_bytes": 0,
		},
		"is_cluster": true,
		"block_upgrade_task": map[string]interface{}{
			"type":     "INVALID",
			"progress": 0,
			"taskId":   "BlockTaskId<0>",
			"state":    "IDLE",
		},
		"grim_reaper": map[string]interface{}{
			"enabled":                    true,
			"is_cluster_fully_connected": true,
			"node_with_least_links":      "",
		},
		"stripe_protection_drives":              2,
		"guid":                                  "9724f5ec-a68c-437d-8411-03c8425c06b8",
		"start_io_starting_drives_grace_secs":   60,
		"start_io_starting_io_nodes_grace_secs": 30,
		"hanging_ios": map[string]interface{}{
			"event_driver_frontend_threshold_secs":                  1800,
			"last_emitted_backend_no_longer_detected_event":         "",
			"alerts_threshold_secs":                                 900,
			"last_emitted_backend_event":                            "",
			"event_backend_threshold_secs":                          1800,
			"last_emitted_driver_frontend_no_longer_detected_event": "",
			"last_emitted_driver_frontend_event":                    "",
			"event_nfs_frontend_threshold_secs":                     1800,
			"last_emitted_nfs_frontend_event":                       "",
			"last_emitted_nfs_frontend_no_longer_detected_event":    "",
		},
		"last_init_failure_time": "",
		"nodes": map[string]interface{}{
			"blacklisted": 0,
			"total":       14,
		},
		"licensing": map[string]interface{}{
			"mode": "Classic",
			"usage": map[string]interface{}{
				"drive_capacity_gb":  3324,
				"usable_capacity_gb": 1829,
				"obs_capacity_gb":    0,
			},
			"license": "license_key",
			"limits": map[string]interface{}{
				"drive_capacity_gb":  3324,
				"usable_capacity_gb": usableCapacity,
				"obs_capacity_gb":    1000000000000000,
				"valid_from":         "2021-03-10T09:46:18Z",
				"expires_at":         "2021-04-09T09:46:18Z",
			},
			"next_check":          "2021-03-10T10:46:20Z",
			"error":               "",
			"check_interval_secs": 3600,
		},
	}

	return ctx.JSON(http.StatusOK, N200{
		Data: &clusterData,
	})
}

func (*MockWekaAPI) Login(ctx v4.Context) error {
	var answer interface{} = map[string]interface{}{
		"access_token":             "eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ",
		"token_type":               "Bearer",
		"expires_in":               3000,
		"refresh_token":            "refresh_token_string",
		"password_change_required": false,
	}

	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (*MockWekaAPI) GetOrganizations(ctx v4.Context) error {
	var answer interface{} = []interface{}{&organization}
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (*MockWekaAPI) CreateOrganization(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &organization,
	})
}

func (*MockWekaAPI) GetMultipleOrgExist(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) DeleteOrganization(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) GetOrganization(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &organization,
	})
}

func (*MockWekaAPI) UpdateOrganization(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &organization,
	})
}

func (*MockWekaAPI) SetOrganizationLimit(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &organization,
	})
}

func (*MockWekaAPI) GetUsers(ctx v4.Context) error {
	var answer interface{} = []interface{}{&user}
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (*MockWekaAPI) CreateUser(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &user,
	})
}

func (*MockWekaAPI) UpdateUserPassword(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) WhoAmI(ctx v4.Context) error {
	var answer interface{} = map[string]interface{}{
		"org_id":   0,
		"username": "user",
		"uid":      "uid_string",
		"source":   "Internal",
		"role":     "ClusterAdmin",
		"org_name": "Root",
	}
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (*MockWekaAPI) DeleteUser(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) UpdateUser(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &user,
	})
}

func (*MockWekaAPI) SetUserPassword(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) RevokeUser(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) GetFileSystemGroups(ctx v4.Context) error {
	var answer interface{} = []interface{}{&fsGroup}
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (*MockWekaAPI) CreateFileSystemGroup(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &fsGroup,
	})
}

func (*MockWekaAPI) DeleteFileSystemGroup(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) GetFileSystemGroup(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &fsGroup,
	})
}

func (*MockWekaAPI) UpdateFileSystemGroup(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &fsGroup,
	})
}

func (*MockWekaAPI) GetFileSystems(ctx v4.Context, _ GetFileSystemsParams) error {
	var answer interface{} = []interface{}{&filesystem}
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (*MockWekaAPI) CreateFileSystem(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &filesystem,
	})
}

func (s *MockWekaAPI) DeleteFileSystem(ctx v4.Context, _ string, _ DeleteFileSystemParams) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.deleteCalled = true
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (s *MockWekaAPI) GetFileSystem(ctx v4.Context, _ string, _ GetFileSystemParams) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deleteCalled {
		s.deleteCalled = false
		return ctx.JSON(http.StatusNotFound, N404{})
	}

	return ctx.JSON(http.StatusOK, N200{
		Data: &filesystem,
	})
}

func (*MockWekaAPI) UpdateFileSystem(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &filesystem,
	})
}

func (*MockWekaAPI) GetS3Buckets(ctx v4.Context) error {
	var answer interface{} = map[string]interface{}{
		"buckets": []interface{}{&bucket},
	}
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (*MockWekaAPI) CreateS3Bucket(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &bucket,
	})
}

func (*MockWekaAPI) DestroyS3Bucket(ctx v4.Context, _ string, _ DestroyS3BucketParams) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) S3ListAllLifecycleRules(ctx v4.Context, _ string) error {
	var answer interface{} = map[string]interface{}{
		"bucket": "bucketName",
		"rules":  []interface{}{&lifecycleRule},
	}
	return ctx.JSON(http.StatusOK, N200{
		Data: &answer,
	})
}

func (*MockWekaAPI) S3CreateLifecycleRule(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &lifecycleRule,
	})
}

func (*MockWekaAPI) S3DeleteLifecycleRule(ctx v4.Context, _ string, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) S3DeleteAllLifecycleRules(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) GetS3BucketPolicy(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &bucketPolicy,
	})
}

func (*MockWekaAPI) SetS3BucketPolicy(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) SetS3BucketQuota(ctx v4.Context, _ string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) CreateS3Policy(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) AttachS3Policy(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) DetachS3Policy(ctx v4.Context) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) DeleteS3Policy(ctx v4.Context, policy string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: nil,
	})
}

func (*MockWekaAPI) GetS3Policy(ctx v4.Context, policy string) error {
	return ctx.JSON(http.StatusOK, N200{
		Data: &s3Policy,
	})
}

func NewMockWekaAPI() *MockWekaAPI {
	return &MockWekaAPI{}
}
