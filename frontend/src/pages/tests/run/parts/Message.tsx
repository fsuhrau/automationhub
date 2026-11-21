import React, {useState} from 'react';
import {Box, Link, Typography} from '@mui/material';
// error text needs the lighter red of the ramp to read on the dark surface
import {red} from '../../../../shared-theme/themePrimitives';
import {splitMessage} from '../logModel';

const lines = (detail: string): string => {
    const count = detail.split('\n').length;
    return `${count} ${count === 1 ? 'line' : 'lines'}`;
};

interface MessageProps {
    text: string,
    error?: boolean,
    open?: boolean,
}

// Message shows the first line of a log message and folds the rest (stack
// traces, JSON) behind "more"
export const Message: React.FC<MessageProps> = ({text, error, open: initiallyOpen}) => {
    const {headline, detail} = splitMessage(text);
    const [open, setOpen] = useState(!!initiallyOpen);
    return <Box sx={{minWidth: 0}}>
        <Typography variant="body2"
            sx={theme => ({
                        fontFamily: 'monospace', fontSize: 13, wordBreak: 'break-word', whiteSpace: 'pre-wrap',
                        color: error ? 'error.main' : 'text.primary',
                        ...(error ? theme.applyStyles('dark', {color: red[300]}) : {}),
                    })}>
            {headline || <Typography component="span" variant="body2" color="text.secondary">(no message)</Typography>}
            {detail && !open && <Link component="button" variant="caption" onClick={() => setOpen(true)} sx={{ml: 1}}>
                more ({lines(detail)})
            </Link>}
        </Typography>
        {detail && open && <>
            <Typography variant="body2" color="text.secondary"
                sx={{fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-word', whiteSpace: 'pre-wrap'}}>
                {detail}
            </Typography>
            <Link component="button" variant="caption" onClick={() => setOpen(false)}>less</Link>
        </>}
    </Box>;
};
