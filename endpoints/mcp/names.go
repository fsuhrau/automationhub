package mcp

import "github.com/fsuhrau/automationhub/storage/models"

// platformNames names the platforms the way the tools take them.
func platformNames() map[models.PlatformType]string {
	return map[models.PlatformType]string{
		models.PlatformTypeiOS:          "ios",
		models.PlatformTypeAndroid:      "android",
		models.PlatformTypeMac:          "mac",
		models.PlatformTypeWindows:      "windows",
		models.PlatformTypeLinux:        "linux",
		models.PlatformTypeWeb:          "web",
		models.PlatformTypeEditor:       "unity_editor",
		models.PlatformTypeiOSSimulator: "ios_simulator",
	}
}

// resultNames names the results of a protocol.
func resultNames() map[models.TestResultState]string {
	return map[models.TestResultState]string{
		models.TestResultOpen:     "open",
		models.TestResultUnstable: "unstable",
		models.TestResultFailed:   "failed",
		models.TestResultSuccess:  "success",
	}
}

// testTypeNames names the types of a test.
func testTypeNames() map[models.TestType]string {
	return map[models.TestType]string{
		models.TestTypeUnity:    "unity",
		models.TestTypeCocos:    "cocos",
		models.TestTypeSerenity: "serenity",
		models.TestTypeScenario: "scenario",
	}
}
