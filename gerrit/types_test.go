package gerrit_test

import (
	"encoding/json"
	"github.com/att-comdev/jarvis-connector/gerrit"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTypes_TimestampToString(t *testing.T) {
	var timestamp = gerrit.Timestamp(time.Now())
	_ = timestamp.String()

	// Test Failure/Error for bad timestamp format
	absolutelyNotATimestamp := []byte("5rNYEA4pMfHwhjc5QBiFiGypB4g4vc")
	err := timestamp.UnmarshalJSON(absolutelyNotATimestamp)
	if err == nil {
		t.Errorf("Expected Unmarshalling to fail and return error")
	}
}

func TestTypes_CheckerInfoToString(t *testing.T) {
	aTimestamp := gerrit.Timestamp(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))
	checkerInfo := gerrit.CheckerInfo{
		UUID: "JarvisLandry",
		URL: "website.com",
		Repository: "repository",
		Blocking: []string{"blocker1", "blocker2"},
		Query: "select * from table",
		Created: aTimestamp,
		Updated: aTimestamp,
	}
	asymmetricCheckerInfo := gerrit.CheckerInfo{
		UUID: "JarvisLandry",
		Name: "Jarvis",
		URL: "website.com",
		Repository: "repository",
		Blocking: []string{"blocker1", "blocker2"},
		Query: "select * from table",
		Created: aTimestamp,
		Updated: aTimestamp,
	}

	checkerInfoString := checkerInfo.String()
	asymmetricCheckerInfoString := asymmetricCheckerInfo.String()

	var checkerInfoUnmarshalled gerrit.CheckerInfo
	var asymmetricCheckerInfoUnmarshalled gerrit.CheckerInfo

	err := json.Unmarshal([]byte(checkerInfoString), &checkerInfoUnmarshalled)
	if err != nil {
		t.Errorf("Could not unmarshal object")
	}
	err = json.Unmarshal([]byte(asymmetricCheckerInfoString), &asymmetricCheckerInfoUnmarshalled)
	if err != nil {
		t.Errorf("Could not unmarshal object")
	}
	if !reflect.DeepEqual(checkerInfo, checkerInfoUnmarshalled) {
		t.Errorf("Marshalled/Unmarshalled object does not match original object")
	}
	if !reflect.DeepEqual(asymmetricCheckerInfo, asymmetricCheckerInfoUnmarshalled) {
		t.Errorf("Marshalled/Unmarshalled object does not match original object")
	}
}

func TestTypes_CheckInputToString(t *testing.T) {
	aTimestamp := gerrit.Timestamp(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))
	checkInput := gerrit.CheckInput{
		CheckerUUID: "JarvisLandry",
		State: "Missouri",
		Message: "This data is made up and doesn't represent realistic values",
		URL: "website.com",
		Started: &aTimestamp,
	}

	checkInputString := checkInput.String()
	var checkInputUnmarshalled gerrit.CheckInput
	err := json.Unmarshal([]byte(checkInputString), &checkInputUnmarshalled)
	if err != nil {
		t.Errorf("Could not unmarshal object")
	}
	if !reflect.DeepEqual(checkInput, checkInputUnmarshalled) {
		t.Errorf("Marshalled/Unmarshalled object does not match original object")
	}
}

func TestTypes_CheckablePatchSetInfo(t *testing.T) {
	checkablePatchSetInfo := gerrit.CheckablePatchSetInfo{
		Repository: "repository",
		ChangeNumber: 1,
		PatchSetID: 123,
	}

	checkablePatchSetInfoString := checkablePatchSetInfo.String()
	var checkablePatchSetInfoUnmarshalled gerrit.CheckablePatchSetInfo

	err := json.Unmarshal([]byte(checkablePatchSetInfoString), &checkablePatchSetInfoUnmarshalled)
	if err != nil {
		t.Errorf("Could not unmarshal object")
	}
	if !reflect.DeepEqual(checkablePatchSetInfo, checkablePatchSetInfoUnmarshalled) {
		t.Errorf("Marshalled/Unmarshalled object does not match original object")
	}
}

func TestTypes_PendingCheckInfo(t *testing.T) {
	pendingCheckInfo := gerrit.PendingCheckInfo{
		State: "Missouri",
	}

	pendingCheckInfoString := pendingCheckInfo.String()
	var pendingCheckInfoUnmarshalled gerrit.PendingCheckInfo

	err := json.Unmarshal([]byte(pendingCheckInfoString), &pendingCheckInfoUnmarshalled)
	if err != nil {
		t.Errorf("Could not unmarshal object")
	}
	if !reflect.DeepEqual(pendingCheckInfo, pendingCheckInfoUnmarshalled) {
		t.Errorf("Marshalled/Unmarshalled object does not match original object")
	}
}

func TestTypes_CheckInput(t *testing.T) {
	aTimestamp := gerrit.Timestamp(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))
	checkInput := gerrit.CheckInput{
		CheckerUUID: "Danny",
		State: "Missouri",
		Message: "Hello World",
		URL: "website.com",
		Started: &aTimestamp,
	}

	checkInputString := checkInput.String()
	var checkInputUnmarshalled gerrit.CheckInput

	err := json.Unmarshal([]byte(checkInputString), &checkInputUnmarshalled)
	if err != nil {
		t.Errorf("Could not unmarshal object")
	}
	if !reflect.DeepEqual(checkInput, checkInputUnmarshalled) {
		t.Errorf("Marshalled/Unmarshalled object does not match original object")
	}
}

func TestTypes_Unmarshal(t *testing.T) {
	longFailure := []byte("x3q8NuFGia3MDDZMmibvyLYHLXZeF5Y57UrD9pcAXSvr5YTUNpZqDusxmcLs9tHXaf5iJ6BXfrnLaEYyHkY76Rrq5ZmbcfsqY6r6tDG6ycNpZ1")
	// a 100 character shorted version of longFailure to ensure error truncation occurs
	oneHundredCharacterLongFailure := "x3q8NuFGia3MDDZMmibvyLYHLXZeF5Y57UrD9pcAXSvr5YTUNpZqDusxmcLs9tHXaf5iJ6BXfrnLaEYyHkY76Rrq5ZmbcfsqY6r"
	shortFailure := []byte("DKqLJwFmktr8RkTAyraqriJTdKgtxkz7BQKNx6Mv")
	myjson := ")]}'\n[]\n"
	validGerritJson := []byte(myjson)

	var longFailureDest interface{}
	var shortFailureDest interface{}
	var validGerritJsonDest interface{}

	longErr := gerrit.Unmarshal(longFailure, &longFailureDest)
	shortErr := gerrit.Unmarshal(shortFailure, &shortFailureDest)
	validErr := gerrit.Unmarshal(validGerritJson, &validGerritJsonDest)

	if longErr != nil && (strings.Contains(longErr.Error(), string(longFailure)) || !strings.Contains(longErr.Error(), oneHundredCharacterLongFailure)) {
		t.Errorf("Error message should contain the trucated first 100 characters of error message")
	}
	if shortErr != nil && !strings.Contains(shortErr.Error(), string(shortFailure)) {
		t.Errorf("Error message should contain entire error when error content is less than 100 characters")
	}
	if validErr != nil {
		t.Errorf("Valid JSON data not parsed successfully")
	}
}
