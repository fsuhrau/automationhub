import {CancelRounded, CheckCircleRounded, ErrorRounded, HourglassEmptyRounded} from '@mui/icons-material';
import {getTestStatusChipColor} from '../../helper/TestStatusHelper';
import {TestResultState} from '../../types/test.result.state.enum';

export type ResultColor = 'success' | 'error' | 'warning' | 'disabled';

export const RESULT_STATES: TestResultState[] = [
    TestResultState.TestResultOpen,
    TestResultState.TestResultUnstable,
    TestResultState.TestResultFailed,
    TestResultState.TestResultSuccess,
];

const byState = <T>(value: (state: TestResultState) => T): Record<TestResultState, T> =>
    Object.fromEntries(RESULT_STATES.map(s => [s, value(s)])) as Record<TestResultState, T>;

// the chip color of TestStatusHelper, with "default" being the icon's "disabled"
export const resultColor = byState<ResultColor>(s => {
    const chip = getTestStatusChipColor(s);
    return chip === 'default' ? 'disabled' : chip;
});

export const resultIcon = {
    [TestResultState.TestResultOpen]: HourglassEmptyRounded,
    [TestResultState.TestResultUnstable]: ErrorRounded,
    [TestResultState.TestResultFailed]: CancelRounded,
    [TestResultState.TestResultSuccess]: CheckCircleRounded,
};
