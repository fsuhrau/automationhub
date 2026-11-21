import React from 'react';
import {Box, CircularProgress, Tooltip} from '@mui/material';
import {getTestStatusText} from '../../helper/TestStatusHelper';
import {TestResultState} from '../../types/test.result.state.enum';
import {resultColor, resultIcon} from './resultIcons';

export interface ResultIconProps {
    result: TestResultState,
    // an open result that is still being produced shows a spinner
    running?: boolean,
}

// ResultIcon shows the state of a test result as an icon with its name as tooltip
export const ResultIcon: React.FC<ResultIconProps> = ({result, running}) => {
    const label = getTestStatusText(result);
    const Icon = resultIcon[result];
    const spinning = running && result === TestResultState.TestResultOpen;
    return <Tooltip title={label}>
        <Box component="span" sx={{display: 'inline-flex'}} aria-label={label}>
            {spinning ? <CircularProgress size={16} color="info"/> : <Icon color={resultColor[result]} fontSize="small"/>}
        </Box>
    </Tooltip>;
};
