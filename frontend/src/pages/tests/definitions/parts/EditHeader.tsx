import React from 'react';
import {Box, Button, CircularProgress, Stack, Tooltip, Typography} from '@mui/material';
import ITestData from '../../../../types/test';
import {TestType} from '../../../../types/test.type.enum';

interface EditHeaderProps {
    test?: ITestData,
    appName?: string,
    dirty: boolean,
    saving: boolean,
    // the save button can't be used: nothing to save or errors to fix
    saveHint: string,
    saveDisabled: boolean,
    onCancel: () => void,
}

const Subtitle: React.FC<Pick<EditHeaderProps, 'test' | 'appName' | 'dirty'>> = ({test, appName, dirty}) =>
    <Typography variant="body2" color="text.secondary">
        {test
            ? <>{TestType[test.testConfig.type]} test · {dirty ? 'unsaved changes' : 'no changes yet'}</>
            : <>A Unity test of {appName ?? 'this app'}: name it, pick what it runs and where.</>}
    </Typography>;

// EditHeader titles the form and has its Cancel and Save buttons
export const EditHeader: React.FC<EditHeaderProps> = ({test, appName, dirty, saving, saveHint, saveDisabled, onCancel}) =>
    <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} useFlexGap={true} alignItems={{sm: 'center'}} sx={{mb: 2}}>
        <Box sx={{minWidth: 0}}>
            <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>
                {test ? `Edit ${test.name}` : 'New test'}
            </Typography>
            <Subtitle test={test} appName={appName} dirty={dirty}/>
        </Box>
        <Stack direction="row" spacing={1} sx={{ml: {sm: 'auto'}}}>
            <Button onClick={onCancel} disabled={saving}>Cancel</Button>
            <Tooltip title={saveHint}>
                <span><Button type="submit" variant="contained" disabled={saveDisabled || saving}
                    startIcon={saving ? <CircularProgress size={16} color="inherit"/> : undefined}>
                    {test ? 'Save' : 'Create test'}
                </Button></span>
            </Tooltip>
        </Stack>
    </Stack>;
