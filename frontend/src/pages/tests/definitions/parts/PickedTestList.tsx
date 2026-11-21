import React from 'react';
import {Box, IconButton, Stack, Tooltip, Typography} from '@mui/material';
import {CloseRounded} from '@mui/icons-material';
import IAppFunctionData from '../../../../types/app.function';
import {functionKey, groupByClass, shortClass, shortMethod} from '../../testEditModel';

interface PickedTestListProps {
    picked: IAppFunctionData[],
    onChange: (picked: IAppFunctionData[]) => void,
    error?: string,
}

const PickedFunction: React.FC<{ fn: IAppFunctionData, onRemove: () => void }> = ({fn, onRemove}) =>
    <Stack direction="row" alignItems="center" sx={{pl: 1}}>
        <Typography variant="body2" sx={{fontFamily: 'monospace', fontSize: 13, flex: 1, wordBreak: 'break-all'}}>{shortMethod(fn.method)}</Typography>
        <Tooltip title="Remove">
            <IconButton size="small" aria-label={`Remove ${shortMethod(fn.method)}`} onClick={onRemove}>
                <CloseRounded fontSize="small"/>
            </IconButton>
        </Tooltip>
    </Stack>;

// PickedTestList shows the picked Unity tests by class and removes them
export const PickedTestList: React.FC<PickedTestListProps> = ({picked, onChange, error}) =>
    <Box>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 0.5}}>Picked tests ({picked.length})</Typography>
        {error && <Typography variant="body2" color="error" sx={{mb: 1}}>{error}</Typography>}
        {picked.length === 0
            ? <Typography variant="body2" color="text.secondary">None yet; add them from the list below.</Typography>
            : <Box sx={{maxHeight: 280, overflowY: 'auto', border: 1, borderColor: 'divider', borderRadius: 1, p: 1}}>
                {groupByClass(picked).map(([cls, fns]) => <Box key={cls} sx={{mb: 1}}>
                    <Typography variant="caption" color="text.secondary" title={cls}>{shortClass(cls)}</Typography>
                    {fns.map(f => <PickedFunction key={functionKey(f)} fn={f}
                        onRemove={() => onChange(picked.filter(p => functionKey(p) !== functionKey(f)))}/>)}
                </Box>)}
            </Box>}
    </Box>;
