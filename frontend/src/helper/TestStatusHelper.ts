import {TestResultState} from "../types/test.result.state.enum";
import {OverridableStringUnion} from "@mui/types";

export const getTestStatusText = (state: TestResultState): string => {
    switch (state) {
        case TestResultState.TestResultSuccess:
            return 'Success';
        case TestResultState.TestResultFailed:
            return 'Failed';
        case TestResultState.TestResultOpen:
            return 'Open';
        case TestResultState.TestResultUnstable:
            return 'Unstable';
    }
    return ''
}

export const getTestStatusColor = (state: TestResultState): string => {
    switch (state) {
        case TestResultState.TestResultSuccess:
            return 'green';
        case TestResultState.TestResultFailed:
            return 'red';
        case TestResultState.TestResultOpen:
            return 'gray';
        case TestResultState.TestResultUnstable:
            return 'yellow';
    }
    return 'green'
}

export const getTestStatusChipColor = (state: TestResultState): any => {
    switch (state) {
        case TestResultState.TestResultSuccess:
            return 'success';
        case TestResultState.TestResultFailed:
            return 'error';
        case TestResultState.TestResultOpen:
            return 'default';
        case TestResultState.TestResultUnstable:
            return 'warning';
    }
    return 'default'
}