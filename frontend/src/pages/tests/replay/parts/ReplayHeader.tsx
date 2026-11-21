import React from 'react';
import {Box, Button, Stack, Typography} from '@mui/material';
import {ArrowBackRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import ITestRunData from '../../../../types/test.run';
import {formatDateTime} from '../../../../components/DateTime';

interface ReplayHeaderProps {
    run: ITestRunData,
    backTo: string,
}

export const ReplayHeader: React.FC<ReplayHeaderProps> = ({run, backTo}) =>
    <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
        <Box sx={{minWidth: 0}}>
            <Typography component="h1" variant="h5">
                {run.test?.name} <Typography component="span" variant="h5" color="text.secondary">· Run {run.id} · Replay</Typography>
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{mt: 0.5}}>
                Started {formatDateTime(run.createdAt, 'YYYY/MM/DD HH:mm')}
            </Typography>
        </Box>
        <Button size="small" startIcon={<ArrowBackRounded/>} component={RouterLink} to={backTo} sx={{ml: {md: 'auto'}}}>
            Back to the run
        </Button>
    </Stack>;
