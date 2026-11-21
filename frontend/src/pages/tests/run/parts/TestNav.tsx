import React from 'react';
import {Button, Stack, Tooltip, Typography} from '@mui/material';
import {ArrowBackRounded, ArrowForwardRounded, SkipNextRounded} from '@mui/icons-material';
import {Neighbours} from '../../protocolModel';
import {TestRow} from '../../testRunModel';

interface TestNavProps {
    place: Neighbours,
    onOpen: (row: TestRow) => void,
}

const NextFailure: React.FC<TestNavProps> = ({place, onOpen}) => {
    const problem = place.nextProblem;
    if (!problem || problem === place.next) return null;
    return <Tooltip title={`Next failed or unstable test: ${problem.name}`}>
        <Button size="small" variant="outlined" color="error" endIcon={<SkipNextRounded/>} onClick={() => onOpen(problem)}>Next failure</Button>
    </Tooltip>;
};

// the way to the other tests of the run
export const TestNav: React.FC<TestNavProps> = (props) => {
    const {place, onOpen} = props;
    const {prev, next} = place;
    return <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {md: 'auto'}, flexWrap: 'wrap'}} useFlexGap={true}
        role="navigation" aria-label="Tests of this run">
        {place.total > 1 && place.position > 0 && <>
            <Tooltip title={prev ? `Previous test: ${prev.name}` : 'This is the first test of the run'}>
                <span><Button size="small" startIcon={<ArrowBackRounded/>} disabled={!prev} onClick={() => prev && onOpen(prev)}>Previous</Button></span>
            </Tooltip>
            <Typography variant="body2" color="text.secondary">Test {place.position} of {place.total}</Typography>
            <Tooltip title={next ? `Next test: ${next.name}` : 'This is the last test of the run'}>
                <span><Button size="small" endIcon={<ArrowForwardRounded/>} disabled={!next} onClick={() => next && onOpen(next)}>Next</Button></span>
            </Tooltip>
            <NextFailure {...props}/>
        </>}
    </Stack>;
};
