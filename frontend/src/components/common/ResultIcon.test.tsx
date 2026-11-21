import {describe, expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';
import {ResultIcon} from './ResultIcon';
import {resultColor, resultIcon} from './resultIcons';
import {TestResultState} from '../../types/test.result.state.enum';

describe('ResultIcon', () => {
    it.each([
        {state: TestResultState.TestResultSuccess, label: 'Success', icon: 'CheckCircleRoundedIcon', color: 'success'},
        {state: TestResultState.TestResultFailed, label: 'Failed', icon: 'CancelRoundedIcon', color: 'error'},
        {state: TestResultState.TestResultUnstable, label: 'Unstable', icon: 'ErrorRoundedIcon', color: 'warning'},
        {state: TestResultState.TestResultOpen, label: 'Open', icon: 'HourglassEmptyRoundedIcon', color: 'disabled'},
    ])('shows state $label', ({state, label, icon, color}) => {
        render(<ResultIcon result={state}/>);
        expect(screen.getByLabelText(label)).toBeInTheDocument();
        expect(screen.getByTestId(icon)).toBeInTheDocument();
        expect(resultColor[state]).toBe(color);
        expect(resultIcon[state]).toBeDefined();
    });

    it('shows a spinner for an open result that is running', () => {
        render(<ResultIcon result={TestResultState.TestResultOpen} running={true}/>);
        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('shows the icon of a finished result even if running', () => {
        render(<ResultIcon result={TestResultState.TestResultFailed} running={true}/>);
        expect(screen.queryByRole('progressbar')).toBeNull();
    });
});
