// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	query "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/query"
	reconciler "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/reconciler_query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	clusterUuidGet   string
	nodegroupUuidGet string
	vipIdGet         int32
)

func insertClusterIntoDb(t *testing.T) (string, error) {
	// Setup mock
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed")
	}
	// THIS FUNCTION ASSUMES THAT CREATE CLUSTER WORKS. REMOVE ONCE WE HAVE DB MIGRATE FOR TESTS SET UP
	res, err := iksSrv.CreateNewCluster(context.Background(), &clusterCreate)
	if err != nil {
		return "", err
	}
	return res.Uuid, nil
}

func insertNodeGroupIntoDB(t *testing.T, tempClusterUuid string) (string, error) {
	// THIS FUNCTION ASSUMES THAT CREATE CLUSTER WORKS. REMOVE ONCE WE HAVE DB MIGRATE FOR TESTS SET UP
	// Initialize compute client mocking
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, nodeGroupCreate.Name, logkeys.CloudAccountId, nodeGroupCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitVnetMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with compute mocking failed")
	}

	//Set up
	original := nodeGroupCreate.Clusteruuid
	nodeGroupCreate.Clusteruuid = tempClusterUuid
	nodeGroupCreate.Instancetypeid = "vm-spr-tny"
	res, err := iksSrv.CreateNodeGroup(context.Background(), &nodeGroupCreate)
	nodeGroupCreate.Clusteruuid = original
	if err != nil {
		return "", err
	}
	return res.Nodegroupuuid, nil
}

func insertVipIntoDB(tempClusterUuid string) (int32, error) {
	// THIS FUNCTION ASSUMES THAT CREATE CLUSTER WORKS. REMOVE ONCE WE HAVE DB MIGRATE FOR TESTS SET UP
	original := vipCreate.Clusteruuid
	vipCreate.Clusteruuid = tempClusterUuid
	res, err := client.CreateNewVip(context.Background(), &vipCreate)
	vipCreate.Clusteruuid = original
	if err != nil {
		return -1, err
	}
	return res.Vipid, nil
}

func insertCertsIntoDbCluster(uuid string, secretKey string) error {
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		return err
	}

	clusterId := 0
	err = sqlDB.QueryRow(query.GetClusterIdQuery, uuid).Scan(&clusterId)
	if err != nil {
		return err
	}
	file, e := os.Create("./encryption_keys")
	if e != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	w.Flush()
	secret := "c3RyaW5nMTIzNA=="
	caCertPlain := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZEekNDQXZlZ0F3SUJBZ0lDQitNd0RRWUpLb1pJaHZjTkFRRUxCUUF3R0RFV01CUUdBMVVFQXhNTmEzVmkKWlhKdVpYUmxjeTFqWVRBZ0dBOHdNREF4TURFd01UQXdNREF3TUZvWERUTXpNRGt5T0RJeE16RTFNMW93R0RFVwpNQlFHQTFVRUF4TU5hM1ZpWlhKdVpYUmxjeTFqWVRDQ0FpSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnSVBBRENDCkFnb0NnZ0lCQU90OWE0UmxUU2lBdDYrQnE1bnZZamttUGhhQyt3Qmh4WHVHQmsrRDRxY0FNeUlUYkt0MnBpQ1AKWUdUUldMT0ZIOXExdTNqS2dZY0JydDlGVVJMNGIyWkJQNDFZdFVTMWJoWFVNSkxROHpWd0xia3NqUi9laUhnTQpqR1AvZk8vcnYwMzFISHFMNUprdlp0Zm92QWxpTWFxdmRqNVplaktQT1MvUzVjRHFFMGRKWERIY2NZZkFtWENUClJFc2JHaFYrS3Rwc0oyMndzTXRHYmJ6RVdoZHFxcm15VE1mM21Ia09ySXkzdmxxZTFhc01VYk5jK2pSUHpTMVYKRXJzOUl0S3pUTXN3WjE3dDRxT3pIMUJubW0zZ0pOMXFwcGlJNGt3Z0tIY3lJODU3bjZCWkp0cVhNRlFmaEtXcAowcXlGUDF3WVk5bWpoejE3UEV2SG9FTWV4ZDVlL1A4N2dqeGEyQUROZWVub053bk5vSHlPWFZXZUliQUpMR2k5ClR0K1haLzhEUldqcmVuYjVIbEZMNEEwb3N0bmh0RUpSaGJvYU50Z0ZEVVBNY0NTYWpXMCt5Q0tkcUZnZUFUYTkKV3d2OUVISmF2K003T2k2T2NwQkxFcmt6bWp4YXZvRVNBK2ROTElGWGlDUHVDYU44bWYxTlc4dWE1ckRXd1VaWApIR3ZYQ0hJMHNaMEdsSWxGRCtBYm1laDZxejFUNnJ5ZzRsUGl4cDFPa2JYNjg1YXhLVlJTd3NQWHdNaUNQR2ZnCmQvNW5BV3UvbWh5YVQraGJ0Q1BnRDFUMVc3Z3Q3eEhLUHdLSmtEbVJhMW1ud1dMUU44Qm1mWHg3NkdQYWFsWFAKdno3d2UyUWNLVG52SjRoZ1o0dzdvbFBad2RpS3FHdEtyZHJrbXExSWl3UmxNQ1ZoY1V6ckFnTUJBQUdqWVRCZgpNQTRHQTFVZER3RUIvd1FFQXdJQ2hEQWRCZ05WSFNVRUZqQVVCZ2dyQmdFRkJRY0RBZ1lJS3dZQkJRVUhBd0V3CkR3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVVE5NEh3R3JZb0hVdUJuTjF6MTdyemMrSDczNHcKRFFZSktvWklodmNOQVFFTEJRQURnZ0lCQUJMTDhMczU2NDVmM2E5SER2STRVNTVOcHhkb2plVnh5VjNtZEFOTgptLytNTXFrTjYybmRQdE5jUEY5cVh5MGRkaTdBeS8wSVV1RnRNd1U5U2Y1ZnZlUXo5cXE1RXc0MG9keTAzNnFUClRrMjlFb25iRzBXcTVQNXdvQnR6TGtaNVRiL1YvVUJOMnlYOGU1NG5jaWdxQ3psUHBIRHMzbkEzZXFlZFVxQmUKTFBDeWJZS2N4ZmdLeTN5am05V2JKV0Nwb2c1SjlrUEVmZnVpTnhFZzJTMVRPdWtZTGpMUnVrNUZYVlBkQW4yZwpsUGVxRVMxd1V1WDJ3bHdwVHo5UUQrR1VHbHI4Y09iYzBWMDJ2YlZhQTllTmVPQitYZjZDVzJDRy9ENWYwbS9VCjRnMGY0cGlVUlFEQWRGSG5hcUhsOWxsczUyZkZFMTY5Sm5xdDlxOTFXZkt5Uy8xTHNSdkIvVmxPeDVZSnhPbEkKTDhZZWxoR0JXaDRzTng4bFVWRmhKRkk0YUNPaWxaWXBESlorb2NORG5VVE9lUHhzY2RjR0dQUXVqVmsrVmN2ZgpuOEJvby8reDdVV3pJSzRqc3VxTWk3bGZIWkxGQWhCdHNBMmJFYkYrK3F2SlZqYk12Y0VyRU9VM0VyRmlKUzZiCnVwNERxSU54bmRsM2NiMmdZVkZjVlVuaUxyWHNEMGRyVXFJUHUxY0RJdkM0bWFaUXVFSzU0aFVBZ1ZzY3hleTcKY3JvRHQ1NlB0ckZvNnhiZkZLY0JBTHhSZGZjRkU3T2VQVE9KL0w5OUZVYVhMQnk0QjJsRERUQUZITm93TjhaZQplTk5BQ1lBM1UwYjlaODYxRmpxSnVnUG9PUjZpMk1HdmNlVGJCM2Rxakg5WmNZeHAvT21NWDErK2xhS0ZXUHVECnJIL3MKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	caKeyPlain := "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKS2dJQkFBS0NBZ0VBNjMxcmhHVk5LSUMzcjRHcm1lOWlPU1krRm9MN0FHSEZlNFlHVDRQaXB3QXpJaE5zCnEzYW1JSTlnWk5GWXM0VWYyclc3ZU1xQmh3R3UzMFZSRXZodlprRS9qVmkxUkxWdUZkUXdrdER6TlhBdHVTeU4KSDk2SWVBeU1ZLzk4Nyt1L1RmVWNlb3ZrbVM5bTEraThDV0l4cXE5MlBsbDZNbzg1TDlMbHdPb1RSMGxjTWR4eApoOENaY0pORVN4c2FGWDRxMm13bmJiQ3d5MFp0dk1SYUYycXF1YkpNeC9lWWVRNnNqTGUrV3A3VnF3eFJzMXo2Ck5FL05MVlVTdXowaTByTk15ekJuWHUzaW83TWZVR2VhYmVBazNXcW1tSWppVENBb2R6SWp6bnVmb0ZrbTJwY3cKVkIrRXBhblNySVUvWEJoajJhT0hQWHM4UzhlZ1F4N0YzbDc4L3p1Q1BGcllBTTE1NmVnM0NjMmdmSTVkVlo0aApzQWtzYUwxTzM1ZG4vd05GYU90NmR2a2VVVXZnRFNpeTJlRzBRbEdGdWhvMjJBVU5ROHh3SkpxTmJUN0lJcDJvCldCNEJOcjFiQy8wUWNscS80enM2TG81eWtFc1N1VE9hUEZxK2dSSUQ1MDBzZ1ZlSUkrNEpvM3laL1UxYnk1cm0Kc05iQlJsY2NhOWNJY2pTeG5RYVVpVVVQNEJ1WjZIcXJQVlBxdktEaVUrTEduVTZSdGZyemxyRXBWRkxDdzlmQQp5SUk4WitCMy9tY0JhNythSEpwUDZGdTBJK0FQVlBWYnVDM3ZFY28vQW9tUU9aRnJXYWZCWXRBM3dHWjlmSHZvClk5cHFWYysvUHZCN1pCd3BPZThuaUdCbmpEdWlVOW5CMklxb2EwcXQydVNhclVpTEJHVXdKV0Z4VE9zQ0F3RUEKQVFLQ0FnRUFndnppbjJSUnhPUEVTTVdTRkRBSnJNeE80T3ErZjNuakJWQ0psaFZBTDdCMndNK1pOTTdzblZQagpSSEVHSytVeTBNOGhscERkOTZEQ0NzTmQwM1dKVVpHZHJodlh5SDQ2MjcySnYwQ201K1NjS2xKVHRaUnN5SW9DClZXVTVzNktvYU02Y3ByWEYxRWQvcHoxM3lxaHFCQTFSY21FSERiU0pGTWIwc1pnQ1hUYTdKNmo5Sll2R1RjNDgKd0tJMG9odnA3bEVXcFhjUkFDRU96VjlTMVkrcG8xMUFSRUsrOXlkb1oyV1Zab2JQUnpPMUJsWURmckdjNlVoWApBUHVDc1R5MnpKY0NDTlc5cVZ6cllDZ2d5RmxUYUMzNHVRUDdER2tlMlI1MHVGZ2ZkR3ZpcHRoT001ek1oZkd1CmsvUVNTRUh3MkpDVG0yay9Jcy9KbjB3d3QvNlNMcjA0ZzBBbVpTWCt0U1doL0w0aGV2N2h6azdabUQyMFJvNmkKOTI4a0NVbmMxY3ZNZ1FXNHg1aFpFOWxnRXp5WGhCUTNhT3FxUFp3dXYrL2N2NUNDeHc2VmxPL0dyblplNmQzdAo4OTNmU3pkczB6R3ZmWHZIaEV5QW5oeDc0U052YTBqblVadW41eTVPck5YcWlQcHRCUlp0dUxUYzhQWFUvNVZlClN0U0owZEdtZk0wYk1oZXdoK2o1U1pwNG5FeDBscjBob0t3MndVbVlOWERJVHNHeXIxRjJIbWhwR2hYMGRWKzIKdXlTVkRGLytvbjRaeDJvL094bEVmdWJzbkRiSnljTW43bUFSYWtJc1lMOWM0TUlWRFordnVvWERkSVhRcVI1Zgo1blh3SXNrQmdROWF5RFJBN1Q5SXRuZjdtd2pvWTlLWVhqdU04VWdoL0dtUUd5SUJickVDZ2dFQkFQM0tFc3hwClkvVXBYR3pyZlBSbXFyRGtjNUU0REc5YmZWSklmZEdFY0hDdHd1YkR2SFhqclBuQUhGcnZXS0haT3UwTXQ5ZXIKcW5CRDNscXFub0NYM0w4YnZpaE1oOGtDTTB2NDlrT3ZTNnQwMTBQaDdhSnozekJBQ0NLVmlwWlh2LzhONDI3QgpWTkV0UHBoZnkvQkFUMmZlc1U5SEdOVGZwdzhkUXMxTWVab21FRExSZmcrbG12Z0pZVXpjNE5McFptQWxXOTJYCmxHdUQycUl1WDVndk1uY0kwYUNia09iMS9RUU9aZW04ZlIvaWE0aFMwM1dJQi9URGJtb1lBMmtFdzhpNVMyUmYKZVJMRW9OZkNjdVQvZldsQTJEWW9Id1JLaEhBeXRFQzYram5DN2lwZFlrZXpsczBDSFp1dng0TjhmQ0pHbGllTgpLTXhBSXhFWHd4aDRTZmNDZ2dFQkFPMktpbUR0cndDbk40bUsvVUFwS3g2Q2FVazhyaGp5M1dGdmV4d1BMWkJZCkk1Ty9PRmRQdW5aRnFoVmhVMG9JY2kzaUMwM3dyaGREWFBGZDBhQ0pkbDJDNE9OeVlmRzNPaTlOWHFsY0pMYWYKMk1FUjFscFA1bWZoek1VTWFHYUJpSHZ1Q0szd096YnF3SEt5OXd0dUxLdmV1T04zcWpDNDF0NkJQcDZYL0N6RQpPQW92eVZ2b2ljOFRLK2JLNXAydDJDK2c3Nm9lakowSytHWXIwV0lIa1BmU0RHZmlvOVNsUFplcE5zU3poOUlNClQxcDFteVJZYlFwRXpuMGpkdkpvUlhXN3ptNmJDdjRwaUJCVE5Qako1Ymw4NVlQNGt0eXdaOVRESVJHTEh4WVIKeVVGdzVUTnNQM24xNDRNUFFCYnJrdzZrc2lvWE40TEZ2aTFCeFJzRzk2MENnZ0VCQVB4MkNtRkI3ZWV4NzNtQwpnTmozVUpHTGtOTkRPRXVHYlpKdS9vcHYveEo0S0V3N0pyejNjZGs0bkh4eFlIQVFrcWZCWVJpd2Ntb2ZlWkFqCjdtenBwUFNQZW9qSUtNTnk1dWlLanlBaHYxcWVibzNlci9CTTZCY3RlMm83N0pOR1UzNDdxS1ZDdVVja2hRSTcKT2JxVG51b3JBNk5qakhZbXpoOGc4cFVib0ZRUnpVZGdVdERwNHRFZk02V3NqQklEa0kzUVhDU0JaMm5VenFkTApEbGxyaWY0VHpjVEJQRklsSGo3c3U4RGFlblkvWE0xTjU0RWhneGlmV0xVOWtoYmtZSWJLblE4S1VueXpFWFhRCnYwN0NRVlYwNWlCcHBRRTF4NjE5SXdiSmVhWUFIY0FUQ1hOZVdZTXl1WlZvTlVhOWpSYmp0UGMvV3Zoa1RQM3gKNzJmbU1WMENnZ0VCQU9zbi9xdk1RVFpGVkI2dTROMVdwQ0EwL0dRTFZWTFBnMGZGSkR2MzdxUjZET1prSkRPVApjMFZJM0FNRWNYN1Y4NnJtbjBoT1h4b2FqdlZIYXBJaDQzTFpjU2JaZ29yWFdCdWgzWGVPQjY1ZmVpWlFNVU1BCjNGaTAyWkd0SWVGd2xKd1RYclpMSDJQVGJDZGdjbDczZC9QQnJvbEpXc3VYQU9nUFROMldHb2g2eS91UnFTWjYKZzRyak1NL3V4L1VMTi95V3R3eFQ1K0pFRFBxZ29FMEtyb0lYaUQ2RHlLcG4xeHkvNEw5RDk4NkNiMEJmTXZIOQpOQjA1VnEzZG9SQ0RGMWhoVHhDQ0hwVFVxcVYycWZjdGNHVjdkbjk2WW5GbGxiUzBZNVZKZzhIR1k5V01IT3NXClc1U3lUc1BkSWhaT3FpdVI4ZXJuUndZSUdxZ1U5enMxWDVVQ2dnRUFRUGNabWw0eStJMFdlM1VFUFczVlV0dkgKVEdadlpqNEpVSkc3dS9oWm5jR0hQZ3RyaWlXWTNDUks1bEJVWnBNbjhRTEMrZnZRSldaMEVBTEE2ZkdCUlJRMwpHT1Y4ZFNNUFZnMzlJM1pwdUxrU25RZzVrZzl2LzhrRWJHQTIzYnU5ODAwOGVHT3RYdllob005cmZTMGxEWHpHCmVUbWVob0ZObmZRRFpMZ2loK1BKbXdWK1hjdjBMRW13MDZjaXNkUlpkYTJ1MlhMY3g1ZFEvU0o4N1paUUpvR3QKbkkvRkI1NnV2RUhVK3VUV3c5Qmc1YXhvUmF1U0IvbEpMK2hLZjJyQWdBVk04cDE2eWZKTlNlbWh0cW5Ud1N0NgpmaXhXbWx5d2VWYXlSdlRNNVVrUnB5REJIaXUrQnJiaFIwWW1WM0g4MnpsUmtEWHhRWENPWUdIMzJFUUROdz09Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	encodedNonce := base64.StdEncoding.EncodeToString(nonce)
	encryptionKeyByte, encryptionKeyId, err := utils.GetLatestEncryptionKey(context.Background(), "./encryption_keys")
	caCert, err := utils.AesEncryptSecret(context.Background(), caCertPlain, encryptionKeyByte, nonce)
	if err != nil {
		return err
	}

	caKey, err := utils.AesEncryptSecret(context.Background(), caKeyPlain, encryptionKeyByte, nonce)
	if err != nil {
		return err
	}
	PutClusterCerts := `
		Update cluster_extraconfig set 
			cluster_id = $1,
			cluster_cacrt = $2,
			cluster_cakey = $3,
			cluster_etcd_cacrt = $4,
			cluster_etcd_cakey = $5,
			cluster_etcd_rotation_keys = $6,
			cluster_sa_pub = $7,
			cluster_sa_key = $8,
			cluster_cp_reg_cmd = $9,
			cluster_wk_reg_cmd = $10,
			encryptionkey_id = $11,
			nonce = $12
	`
	_, err = sqlDB.Exec(PutClusterCerts,
		clusterId,
		caCert,
		caKey,
		secret,
		secret,
		secret,
		secret,
		secret,
		secret,
		secret,
		encryptionKeyId,
		encodedNonce,
	)
	if err != nil {
		return err
	}

	err = setActionableState(clusterUuidGet)
	if err != nil {
		return err
	}
	nuuid := uuid
	nuuid = strings.Replace(nuuid, "cl-", "cp-", 1)
	imi := "iks-vm-u22-cd-cp-1-27-11-v20240227"
	clusterStatusStr := `{
    "uuid": "` + uuid + `",
    "clusterStatus": {
    "lastUpdate": "2023-09-28T17:51:19Z", 
    "ilbs": [
      {
        "state": "Active", 
        "name": "` + uuid + `-apiserver", 
        "poolID": 125693, 
        "message": "", 
        "conditions": {
          "vipPoolLinked": true, 
          "poolCreated": true, 
          "vipCreated": true
        }, 
        "vip": "10.10.10.01", 
        "vipID": 132502
      }, 
      {
        "state": "Active", 
        "name": "` + uuid + `-etcd", 
        "poolID": 125692, 
        "message": "", 
        "conditions": {
          "vipPoolLinked": true, 
          "poolCreated": true, 
          "vipCreated": true
        }, 
        "vip": "10.10.10.02",
        "vipID": 132501
      }, 
      {
        "state": "Active", 
        "name": "` + uuid + `-konnectivity", 
        "poolID": 125695, 
        "message": "", 
        "conditions": {
          "vipPoolLinked": true, 
          "poolCreated": true, 
          "vipCreated": true
        }, 
        "vip": "10.10.10.03",
        "vipID": 132504
      }, 
      {
        "state": "Active", 
        "name": "` + uuid + `-public-apiserver", 
        "poolID": 125694, 
        "message": "", 
        "conditions": {
          "vipPoolLinked": true, 
          "poolCreated": true, 
          "vipCreated": true
        }, 
        "vip": "146.152.227.14",
        "vipID": 132503
      }
    ], 
    "nodegroups": [
      {
        "count": 3, 
        "state": "Active", 
        "nodes": [
          {
            "lastUpdate": "2023-09-28T15:46:05Z", 
            "instanceIMI": "` + imi + `",
            "kubeletVersion": "", 
            "name": "` + nuuid + `-cf-4063d", 
            "creationTime": "2023-09-28T15:41:58Z", 
            "reason": "", 
            "kubeProxyVersion": "", 
            "state": "Active", 
            "dnsName": "", 
            "message": "", 
            "ipAddress": "100.80.145.133"
          }, 
          {
            "lastUpdate": "2023-09-28T15:46:05Z", 
            "instanceIMI": "` + imi + `",
            "kubeletVersion": "", 
            "name": "` + nuuid + `-cf-852bd", 
            "creationTime": "2023-09-28T15:44:18Z", 
            "reason": "", 
            "kubeProxyVersion": "", 
            "state": "Active", 
            "dnsName": "", 
            "message": "", 
            "ipAddress": "100.80.145.11"
          }, 
          {
            "lastUpdate": "2023-09-28T15:46:05Z", 
            "instanceIMI": "` + imi + `",
            "kubeletVersion": "", 
            "name": "` + nuuid + `-c8108", 
            "creationTime": "2023-09-28T15:40:14Z", 
            "reason": "", 
            "kubeProxyVersion": "", 
            "state": "Active", 
            "dnsName": "", 
            "message": "", 
            "ipAddress": "100.80.145.226"
          }
        ], 
        "name": "` + nuuid + `", 
        "message": "Nodegroup ready", 
        "type": "controlplane", 
        "reason": ""
      }
    ], 
    "reason": "", 
    "message": "Cluster ready", 
    "addons": [
      {
        "lastUpdate": "2023-09-28T15:46:08Z", 
        "state": "Active", 
        "name": "` + uuid + `-calico-config", 
        "message": "", 
        "reason": "", 
        "artifact": "http://100.64.16.76:7071/calico-config-3260-k1271-1.template"
      }, 
      {
        "lastUpdate": "2023-09-28T15:46:05Z", 
        "state": "Active", 
        "name": "` + uuid + `-calico-operator", 
        "message": "", 
        "reason": "", 
        "artifact": "http://100.64.16.76:7071/calico-operator-3260-k1271-1.template"
      }, 
      {
        "lastUpdate": "2023-09-28T15:46:05Z", 
        "state": "Active", 
        "name": "` + uuid + `-cf-coredns", 
        "message": "", 
        "reason": "", 
        "artifact": "http://100.64.16.76:7071/coredns-171-k1271-1.template"
      }, 
      {
        "lastUpdate": "2023-09-28T15:46:08Z", 
        "state": "Active", 
        "name": "` + uuid + `-konnectivity-agent", 
        "message": "", 
        "reason": "", 
        "artifact": "http://100.64.16.76:7071/konnectivity-agent.template"
      }, 
      {
        "lastUpdate": "2023-09-28T15:46:04Z", 
        "state": "Active", 
        "name": "` + uuid + `-kube-proxy", 
        "message": "", 
        "reason": "", 
        "artifact": "http://100.64.16.76:7071/kube-proxy-k1271-1.template"
      }
    ], 
    "state": "Active"
  }
}`
	var clusterStatus pb.UpdateClusterStatusRequest
	err = json.Unmarshal([]byte(clusterStatusStr), &clusterStatus)
	if err != nil {
		return err
	}
	_, err = reconciler.PutClusterStatusReconcilerQuery(context.Background(), sqlDB, &clusterStatus)
	if err != nil {
		return err
	}

	return nil
}

/* CLUSTER UNIT TESTS */
func TestSetupForClusterGetTests(t *testing.T) {
	var err error
	clearDB()
	// Insert Cluster
	clusterUuidGet, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	// Insert SetActionable
	err = setActionableState(clusterUuidGet)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodegroupUuidGet, err = insertNodeGroupIntoDB(t, clusterUuidGet)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
	// Insert SetActionable
	err = setActionableState(clusterUuidGet)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Vip
	vipIdGet, err = insertVipIntoDB(clusterUuidGet)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
}

func TestGetAllClustersIksUser(t *testing.T) {
	// Set up
	iksCloudAccountId := &pb.IksCloudAccountId{
		CloudAccountId: "iks_user",
	}
	// Test
	res, err := client.GetClusters(context.Background(), iksCloudAccountId)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	if len(res.Clusters) != 1 {
		t.Fatalf("Expected number of clusters not returned, got %d, expected 1", len(res.Clusters))
	}
	if res.Clusters[0].Uuid != clusterUuidGet {
		t.Fatalf("Cluster returns not expected")

	}
}
func TestGetAllClustersRkeUser(t *testing.T) {
	// Set up
	iksCloudAccountId := &pb.IksCloudAccountId{
		CloudAccountId: "rke_user",
	}
	// Test
	res, err := client.GetClusters(context.Background(), iksCloudAccountId)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	if len(res.Clusters) != 0 {
		t.Fatalf("Expected number of clusters not returned")
	}
}

func TestGetClusterExistance(t *testing.T) {
	// Set up
	clusterId := &pb.ClusterID{
		Clusteruuid:    "fake_id",
		CloudAccountId: "iks_user",
	}
	// Test
	_, err := client.GetCluster(context.Background(), clusterId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + clusterId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}

func TestGetClusterPermissions(t *testing.T) {
	// Set up
	clusterId := &pb.ClusterID{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "fake_user",
	}
	// Test
	_, err := client.GetCluster(context.Background(), clusterId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + clusterId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}
func TestGetClusterSuccess(t *testing.T) {
	// Set up
	clusterId := &pb.ClusterID{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "iks_user",
	}
	// Test
	res, err := client.GetCluster(context.Background(), clusterId)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	if (res.Uuid) != clusterId.Clusteruuid {
		t.Fatalf("Result does not match expectation")
	}
	if len(res.Nodegroups) != 1 || res.Nodegroups[0].Clusteruuid != clusterId.Clusteruuid {
		t.Fatalf("Result does not match expectation")
	}
	if len(res.Vips) != 1 {
		t.Fatalf("Result does not match expectation")
	}
}

func TestGetClusterStatusExistance(t *testing.T) {
	// Set up
	clusterId := &pb.ClusterID{
		Clusteruuid:    "fake_id",
		CloudAccountId: "iks_user",
	}
	// Test
	_, err := client.GetClusterStatus(context.Background(), clusterId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + clusterId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}

func TestGetClusterStatusPermissions(t *testing.T) {
	// Set up
	clusterId := &pb.ClusterID{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "fake_user",
	}
	// Test
	_, err := client.GetClusterStatus(context.Background(), clusterId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + clusterId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}

func TestGetClusterStatusSuccess(t *testing.T) {
	// Set up
	clusterId := &pb.ClusterID{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "iks_user",
	}
	// Test
	res, err := client.GetClusterStatus(context.Background(), clusterId)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	if (res.Clusteruuid) != clusterId.Clusteruuid {
		t.Fatalf("Result does not match expectation")
	}
}

func TestGetNodeGroupsExistance(t *testing.T) {
	// Set up
	getNodes := true
	getNodeGroup := &pb.GetNodeGroupRequest{
		Clusteruuid:    "fake_id",
		Nodegroupuuid:  nodegroupUuidGet,
		CloudAccountId: "iks_user",
		Nodes:          &getNodes,
	}
	// Test
	_, err := client.GetNodeGroup(context.Background(), getNodeGroup)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "NodeGroup not found in Cluster: " + getNodeGroup.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}

func TestGetNodeGroupPermissions(t *testing.T) {
	// Set up
	getNodes := true
	getNodeGroup := &pb.GetNodeGroupRequest{
		Clusteruuid:    clusterUuidGet,
		Nodegroupuuid:  nodegroupUuidGet,
		CloudAccountId: "fake_user",
		Nodes:          &getNodes,
	}
	// Test
	_, err := client.GetNodeGroup(context.Background(), getNodeGroup)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + getNodeGroup.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}
func TestGetNodeGroupSuccess(t *testing.T) {
	// Set up
	getNodes := true
	getNodeGroup := &pb.GetNodeGroupRequest{
		Clusteruuid:    clusterUuidGet,
		Nodegroupuuid:  nodegroupUuidGet,
		CloudAccountId: "iks_user",
		Nodes:          &getNodes,
	}
	// Test
	res, err := client.GetNodeGroup(context.Background(), getNodeGroup)
	if err != nil {
		t.Fatalf("Should get not error, received: %v", err.Error())
	}
	// Check
	if res.Nodegroupuuid != getNodeGroup.Nodegroupuuid {
		t.Fatalf("Result does not match expectation")
	}
}
func TestGetNodeGroupExistance(t *testing.T) {
	// Set up
	getNodeGroups := &pb.GetNodeGroupsRequest{
		Clusteruuid:    "fake_id",
		CloudAccountId: "iks_user",
	}
	// Test
	_, err := client.GetNodeGroups(context.Background(), getNodeGroups)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + getNodeGroups.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}

func TestGetNodeGroupsPermissions(t *testing.T) {
	// Set up
	getNodeGroups := &pb.GetNodeGroupsRequest{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "fake_user",
	}
	// Test
	_, err := client.GetNodeGroups(context.Background(), getNodeGroups)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + getNodeGroups.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}
func TestGetNodeGroupsSuccess(t *testing.T) {
	// Set up
	getNodeGroups := &pb.GetNodeGroupsRequest{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "iks_user",
	}
	// Test
	res, err := client.GetNodeGroups(context.Background(), getNodeGroups)
	if err != nil {
		t.Fatalf("Should get not error, received: %v", err.Error())
	}
	// Check
	if len(res.Nodegroups) != 1 {
		t.Fatalf("Result does not match expectation")
	}
}

func TestGetNodegroupStatusExistance(t *testing.T) {
	// Set up
	nodeGroupId := &pb.NodeGroupid{
		Clusteruuid:    "fake_id",
		CloudAccountId: "iks_user",
		Nodegroupuuid:  nodegroupUuidGet,
	}
	// Test
	_, err := client.GetNodeGroupStatus(context.Background(), nodeGroupId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "NodeGroup not found in Cluster: " + nodeGroupId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}
func TestGetNodegroupStatusPermissions(t *testing.T) {
	// Set up
	nodeGroupId := &pb.NodeGroupid{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "fake_user",
		Nodegroupuuid:  nodegroupUuidGet,
	}
	// Test
	_, err := client.GetNodeGroupStatus(context.Background(), nodeGroupId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + nodeGroupId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}
func TestGetNodegroupStatusSuccess(t *testing.T) {
	// Set up
	nodeGroupId := &pb.NodeGroupid{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "iks_user",
		Nodegroupuuid:  nodegroupUuidGet,
	}
	// Test
	res, err := client.GetNodeGroupStatus(context.Background(), nodeGroupId)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	if res.Nodegroupuuid != nodeGroupId.Nodegroupuuid {
		t.Fatalf("Result does not mayytch expectation")
	}

}

func TestGetVipsExistance(t *testing.T) {
	// Set up
	getVips := &pb.ClusterID{
		Clusteruuid:    "fake_id",
		CloudAccountId: "iks_user",
	}
	// Test
	_, err := client.GetVips(context.Background(), getVips)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + getVips.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}

func TestGetVipsPermissions(t *testing.T) {
	// Set up
	getVips := &pb.ClusterID{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "fake_user",
	}
	// Test
	_, err := client.GetVips(context.Background(), getVips)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + getVips.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}

func TestGetVipsSuccess(t *testing.T) {
	// Set up
	getVips := &pb.ClusterID{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "iks_user",
	}
	// Test
	res, err := client.GetVips(context.Background(), getVips)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	if len(res.Response) != 1 {
		t.Fatalf("Result does not mayytch expectation")
	}
}

func TestGetVipExistence(t *testing.T) {
	// Set up
	getVip := &pb.VipId{
		Clusteruuid:    "fake_id",
		Vipid:          vipIdGet,
		CloudAccountId: "iks_user",
	}
	// Test
	_, err := client.GetVip(context.Background(), getVip)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Vip not found in Cluster: " + getVip.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}

func TestGetVipPermissions(t *testing.T) {
	// Set up
	getVip := &pb.VipId{
		Clusteruuid:    clusterUuidGet,
		Vipid:          vipIdGet,
		CloudAccountId: "fake_user",
	}
	// Test
	_, err := client.GetVip(context.Background(), getVip)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + getVip.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}

func TestGetVipSuccess(t *testing.T) {
	// Set up
	getVip := &pb.VipId{
		Clusteruuid:    clusterUuidGet,
		Vipid:          vipIdGet,
		CloudAccountId: "iks_user",
	}
	// Test
	res, err := client.GetVip(context.Background(), getVip)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	if res.Vipid != getVip.Vipid {
		t.Fatalf("Result does not mayytch expectation")
	}
}

func TestGetClustersDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidGet)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

/* GET PUBLIC METADATA INSTANCE TYPES */
func TestGetInstanceTypesForIksUser(t *testing.T) {
	// Set Vars
	errorBase := "Get InstanceTypes for IKS user: "
	expectedInstanceTypes := 4
	cloudAccount := &pb.IksCloudAccountId{
		CloudAccountId: "iks_user",
	}
	// Call Function
	res, err := client.GetPublicInstanceTypes(context.Background(), cloudAccount)
	if err != nil {
		t.Fatalf(errorBase+"%v", err)
	}
	// Check expected
	if len(res.Instancetypes) != expectedInstanceTypes {
		t.Fatalf(errorBase + "Returned instance types do not match expected")
	}
}
func TestGetInstanceTypesForRkeUser(t *testing.T) {
	// Set Vars
	errorBase := "Get InstanceTypes for RKE user: "
	expectedInstanceTypes := 4
	cloudAccount := &pb.IksCloudAccountId{
		CloudAccountId: "rke2_user",
	}
	// Call Function
	res, err := client.GetPublicInstanceTypes(context.Background(), cloudAccount)
	if err != nil {
		t.Fatalf(errorBase+"%v", err)
	}
	// Check expected
	if len(res.Instancetypes) != expectedInstanceTypes {
		t.Fatalf(errorBase + "Returned instance types do not match expected")
	}
}

/* GET PUBLIC METADATA RUNTIME */
func TestGetRuntimesForIksUser(t *testing.T) {
	// Set Vars
	errorBase := "Get Runtimes for IKS user: "
	expectedRuntimes := 1
	expectedK8sversions := 3
	cloudAccount := &pb.IksCloudAccountId{
		CloudAccountId: "iks_user",
	}
	// Call Function
	res, err := client.GetPublicRuntimes(context.Background(), cloudAccount)
	if err != nil {
		t.Fatalf(errorBase+"%v", err)
	}
	// Check expected
	if len(res.Runtimes) != expectedRuntimes {
		t.Fatalf(errorBase + "Returned runtimes do not match expected")
	}
	if len(res.Runtimes[0].K8Sversionname) != expectedK8sversions {
		t.Fatalf(errorBase + "Returned k8sversions do not match expected")
	}
}

/*
func TestGetRuntimesForRkeUser(t *testing.T) {
	// Set Vars
	errorBaseDefault := "Get Runtimes for RKE user: "
	expectedRuntimes := 1
	expectedK8sversions := 2
	cloudAccount := &pb.IksCloudAccountId{
		CloudAccountId: "rke2_user",
	}
	// Call Function
	res, err := client.GetPublicRuntimes(context.Background(), cloudAccount)
	if err != nil {
		t.Fatalf(errorBaseDefault+"%v", err)
	}
	// Check expected
	if len(res.Runtimes) != expectedRuntimes {
		t.Fatalf(errorBaseDefault + "Returned runtimes do not match expected")
	}
	if len(res.Runtimes[0].K8Sversionname) != expectedK8sversions {
		t.Fatalf(errorBaseDefault + "Returned k8sversions do not match expected")
	}
}
*/

/* GET PUBLIC METADATA K8SVERSION NAME */
func TestGetK8sversionsForIksUser(t *testing.T) {
	// Set Vars
	errorBase := "Get K8sversion for IKS user: "
	expectedAmount := 3
	expectedRegex := regexp.MustCompile(`^[0-9]{1}\.[0-9]+$`) //1.27
	cloudAccount := &pb.IksCloudAccountId{
		CloudAccountId: "iks_user",
	}
	// Call Function
	res, err := client.GetPublicK8SVersions(context.Background(), cloudAccount)
	if err != nil {
		t.Fatalf(errorBase+"%v", err)
	}
	// Check expected
	if len(res.K8Sversions) != expectedAmount {
		t.Fatalf(errorBase + "Returned values do not match expected")
	}
	if !expectedRegex.MatchString(res.K8Sversions[0].K8Sversionname) {
		t.Fatalf(errorBase + "Returned K8svesion do not match expected format")
	}
}

/*
func TestGetK8sversionsForRkeUser(t *testing.T) {
	// Set Vars
	errorBaseDefault := "Get K8sversion for rke2 user: "
	expectedAmount := 3
	expectedRegex := regexp.MustCompile(`^v[0-9]{1}(\.[0-9]+){2}(\+rke2r1)$`) //v1.27.2+rke2r1
	cloudAccount := &pb.IksCloudAccountId{
		CloudAccountId: "rke2_user",
	}
	// Call Function
	res, err := client.GetPublicK8SVersions(context.Background(), cloudAccount)
	if err != nil {
		t.Fatalf(errorBaseDefault+"%v", err)
	}
	// Check expected
	if len(res.K8Sversions) != expectedAmount {
		t.Fatalf(errorBaseDefault + "Returned values do not match expected")
	}
	if !expectedRegex.MatchString(res.K8Sversions[0].K8Sversionname) {
		t.Fatalf(errorBaseDefault + "Returned K8svesion do not match expected format")
	}
}
*/

func TestGetKubeConfig(t *testing.T) {
	// Set Up
	var err error
	clusterUuidGet, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	secretKey := "./encryption_keys"
	err = insertCertsIntoDbCluster(clusterUuidGet, secretKey)
	if err != nil {
		t.Fatalf("Could not set up clusterCerts: %v", err)
	}
	request := pb.GetKubeconfigRequest{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: clusterCreate.CloudAccountId,
	}
	// Test
	res, err := query.GetKubeConfig(context.Background(), sqlDb, &request, secretKey)
	if err != nil {
		t.Fatalf("Could not get kubeconfig : %v", err)
	}
	// Check
	if res.Kubeconfig == "" {
		t.Fatalf("Kubeconfig is empty")
	}
}

func TestGetKubeConfigFailPermissions(t *testing.T) {
	// Set up
	request := pb.GetKubeconfigRequest{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: "fake_user",
	}
	secretKey := "./encryption_keys"
	_, err := query.GetKubeConfig(context.Background(), sqlDb, &request, secretKey)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	expectedError := errorBaseNotFound + "Cluster not found: " + request.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Get Admin config: Expected not found error, received: %v", err.Error())
	}
}

func TestGetSecurityRules(t *testing.T) {
	// Set up
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidGet,
		CloudAccountId: clusterCreate.CloudAccountId,
	}
	_, err := client.GetFirewallRule(context.Background(), &clusterId)
	if err != nil {
		t.Fatalf("Should get error, none was received")
	}
}

func TestGetDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidGet)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}
