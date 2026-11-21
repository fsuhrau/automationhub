import React from 'react';
import {Box, CircularProgress, IconButton, Stack, Tooltip, Typography} from '@mui/material';
import {RefreshRounded} from '@mui/icons-material';
import IAppFunctionData from '../../../../types/app.function';
import {addFunctions} from '../../testEditModel';
import {UnityTests, useUnityTests} from '../useUnityTests';
import {PasteTestList} from './PasteTestList';
import {PickedTestList} from './PickedTestList';
import {UnityTestChooser} from './UnityTestChooser';

interface TestPickerProps {
    picked: IAppFunctionData[],
    onChange: (functions: IAppFunctionData[]) => void,
    error?: string,
}

const Asking: React.FC = () =>
    <Stack direction="row" spacing={1} alignItems="center"><CircularProgress size={16}/>
        <Typography variant="body2" color="text.secondary">Asking the Unity editor on this computer for its tests…</Typography></Stack>;

const AddTestsHeader: React.FC<Pick<UnityTests, 'loading' | 'reload'>> = ({loading, reload}) =>
    <Stack direction="row" alignItems="center" spacing={1} sx={{mb: 1}}>
        <Typography variant="body2" sx={{fontWeight: 600}}>Add tests</Typography>
        <Tooltip title="Ask the Unity editor again">
            <span><IconButton size="small" aria-label="Load the tests from Unity again" onClick={reload} disabled={loading}>
                <RefreshRounded fontSize="small"/>
            </IconButton></span>
        </Tooltip>
    </Stack>;

// the pasted list or Unity's, or the way to paste one when Unity is silent
const Source: React.FC<{ unity: UnityTests, picked: IAppFunctionData[], onAdd: (f: IAppFunctionData[]) => void }> = ({unity, picked, onAdd}) => {
    const {available, loading, unreachable, paste} = unity;
    if (loading) return <Asking/>;
    if (available === null || (unreachable && available.length === 0)) return <PasteTestList onUse={paste}/>;
    return <UnityTestChooser available={available} picked={picked} onAdd={onAdd}/>;
};

// TestPicker shows the picked Unity tests and adds more from the Unity
// editor on this computer, or from a pasted list
export const TestPicker: React.FC<TestPickerProps> = ({picked, onChange, error}) => {
    const unity = useUnityTests();
    return <Stack spacing={2}>
        <PickedTestList picked={picked} onChange={onChange} error={error}/>
        <Box>
            <AddTestsHeader loading={unity.loading} reload={unity.reload}/>
            <Source unity={unity} picked={picked} onAdd={list => onChange(addFunctions(picked, list))}/>
        </Box>
    </Stack>;
};
