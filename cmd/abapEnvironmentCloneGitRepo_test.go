package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

var executionLogStringClone string

func init() {
	executionLog := abaputils.PullEntity{
		ToExecutionLog: abaputils.AbapLogs{
			Results: []abaputils.LogResults{
				{
					Index:       "1",
					Type:        "LogEntry",
					Description: "S",
					Timestamp:   "/Date(1644332299000+0000)/",
				},
			},
		},
	}
	executionLogResponse, _ := json.Marshal(executionLog)
	executionLogStringClone = string(executionLogResponse)
}

func TestCloneStep(t *testing.T) {
	t.Run("Run Step - Successful", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir, errDir := ioutil.TempDir("", "test read addon descriptor")
		if errDir != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		body := `---
repositories:
- name: /DMO/REPO_A
  tag: v-1.0.1-build-0001
  branch: branchA
  version: 1.0.1
- name: /DMO/REPO_B
  tag: rel-2.1.1-build-0001
  branch: branchB
  version: 2.1.1
`
		file, _ := os.Create("filename.yaml")
		file.Write([]byte(body))

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch1",
			Repositories:      "filename.yaml",
		}

		logResultSuccess := fmt.Sprintf(`{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Success", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`)
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringClone + `}`,
				logResultSuccess,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : ` + executionLogStringClone + `}`,
				logResultSuccess,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : ` + executionLogStringClone + `}`,
				logResultSuccess,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentCloneGitRepo(&config, &autils, client)
		assert.NoError(t, err, "Did not expect error")
		assert.Equal(t, 0, len(client.BodyList), "Not all requests were done")
	})

	t.Run("Run Step - failing", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch1",
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : {} }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentCloneGitRepo(&config, &autils, client)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Clone of repository / software component 'testRepo1', branch 'testBranch1' failed on the ABAP system: Request to ABAP System not successful", err.Error(), "Expected different error message")
		}

	})
}

func TestCloneStepErrorMessages(t *testing.T) {
	t.Run("Status Error", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir, errDir := ioutil.TempDir("", "test read addon descriptor")
		if errDir != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		body := `---
repositories:
- name: /DMO/REPO_A
  tag: v-1.0.1-build-0001
  branch: branchA
  version: 1.0.1
  commitID: ABCD1234
`
		file, _ := os.Create("filename.yaml")
		file.Write([]byte(body))

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "filename.yaml",
		}

		logResultError := fmt.Sprintf(`{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Error", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`)
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringClone + `}`,
				logResultError,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentCloneGitRepo(&config, &autils, client)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Clone of repository / software component '/DMO/REPO_A', branch 'branchA', commit 'ABCD1234' failed on the ABAP System", err.Error(), "Expected different error message")
		}
	})

	t.Run("Poll Request Error", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch1",
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : {  } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentCloneGitRepo(&config, &autils, client)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Clone of repository / software component 'testRepo1', branch 'testBranch1' failed on the ABAP system: Request to ABAP System not successful", err.Error(), "Expected different error message")
		}
	})

	t.Run("Trigger Clone Error", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch1",
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : {  } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentCloneGitRepo(&config, &autils, client)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Clone of repository / software component 'testRepo1', branch 'testBranch1' failed on the ABAP system: Request to ABAP System not successful", err.Error(), "Expected different error message")
		}
	})

	t.Run("Missing file error", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch1",
			Repositories:      "filename.yaml",
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : {} }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		err := runAbapEnvironmentCloneGitRepo(&config, &autils, client)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Something failed during the clone: Could not find filename.yaml", err.Error(), "Expected different error message")
		}

	})
}
