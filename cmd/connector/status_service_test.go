package main_test

import (
	"github.com/att-comdev/jarvis-connector/cmd/connector"
	"testing"
)

type statusServiceTestData struct {
	input int
	expected string
	result string
}

type statusServiceMock struct {
	stringFn func() string
}

func (mock statusServiceMock) String() string {
	return mock.stringFn()
}

func TestStringMockability(t *testing.T) {
	serviceMock := statusServiceMock{}
	serviceMock.stringFn = func() string {
		return "nonsense"
	}

	// Test to make sure the mocking works
	result := serviceMock.String()
	expected := "nonsense"
	if result != expected {
		t.Error("mock was not called, expected mocked service to return \"nonsense\" string")
	}
}

func TestString(t *testing.T) {
	testData := []statusServiceTestData{
		{
			input:    main.Unset,
			expected: main.UnsetString,
			result:   "",
		}, {
			input:    main.Irrelevant,
			expected: main.IrrelevantString,
			result:   "",
		}, {
			input:    main.Running,
			expected: main.RunningString,
			result:   "",
		}, {
			input:    main.Fail,
			expected: main.FailString,
			result:   "",
		}, {
			input:    main.Successful,
			expected: main.SuccessfulString,
			result:   "",
		},
	}

	// Test to make sure the mapping works
	for _, test := range testData {
		status := main.StatusServiceImpl{Status: test.input}
		test.result = status.String()
		if test.expected != test.result {
			t.Errorf("Test input: %d does not produce expected mapping expected: %s result: %s", test.input, test.expected, test.result)
		}
	}
}
