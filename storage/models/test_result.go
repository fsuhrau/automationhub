package models

type TestResultState uint

const (
	TestResultOpen TestResultState = iota
	TestResultUnstable
	TestResultFailed
	TestResultSuccess
)
