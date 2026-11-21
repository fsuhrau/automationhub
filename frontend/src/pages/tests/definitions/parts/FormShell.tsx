import React from 'react';
import {Box} from '@mui/material';

interface FormShellProps {
    onSubmit: () => void,
    children: React.ReactNode,
}

// FormShell is the edit form: Enter and Cmd/Ctrl+S submit it
export const FormShell: React.FC<FormShellProps> = ({onSubmit, children}) =>
    <Box component="form" noValidate={true} sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2, '& > *': {maxWidth: 960}}}
        onSubmit={e => {
             e.preventDefault();
             onSubmit();
         }}
        onKeyDown={e => {
             if ((e.metaKey || e.ctrlKey) && e.key === 's') {
                 e.preventDefault();
                 onSubmit();
             }
         }}>
        {children}
    </Box>;
