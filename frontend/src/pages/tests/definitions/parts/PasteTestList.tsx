import React, {useState} from 'react';
import {Alert, Box, Button, Stack, TextField} from '@mui/material';
import IAppFunctionData from '../../../../types/app.function';
import {parsePastedTests} from '../../testEditModel';

// PasteTestList is the way out when the Unity editor doesn't answer: the
// test list copied there can be pasted
export const PasteTestList: React.FC<{ onUse: (functions: IAppFunctionData[]) => void }> = ({onUse}) => {
    const [text, setText] = useState('');
    const [error, setError] = useState<string | null>(null);
    const use = (): void => {
        const result = parsePastedTests(text);
        if ('error' in result) {
            setError(result.error);
            return;
        }
        setError(null);
        onUse(result.functions);
    };
    return <Stack spacing={1}>
        <Alert severity="info">
            The Unity editor on this computer didn't answer. Open the project in Unity with the Automation Hub plugin
            (Tools › Automation Hub › Settings) and try again, or copy the test list there and paste it below.
        </Alert>
        <TextField multiline={true} minRows={3} maxRows={8} fullWidth={true} size="small" placeholder='[{"assembly": "…", "class": "…", "method": "…"}]'
            value={text} onChange={e => setText(e.target.value)} error={!!error} helperText={error ?? undefined}
            slotProps={{htmlInput: {'aria-label': 'Pasted test list', spellCheck: false}}}/>
        <Box><Button size="small" variant="outlined" disabled={!text.trim()} onClick={use}>Use this list</Button></Box>
    </Stack>;
};
