import React, {useState} from 'react';
import {IconButton, InputAdornment, TextField} from '@mui/material';
import {SendRounded} from '@mui/icons-material';
import {DeviceInput} from '../deviceInput';

// text to type on the device in one go, e.g. pasted
const TypeField: React.FC<{ input: DeviceInput, disabled: boolean }> = ({input, disabled}) => {
    const [text, setText] = useState('');
    const send = (): void => input.typeText(text, () => setText(''));
    return (
        <TextField size="small" fullWidth={true} placeholder="Type or paste text, then Enter" value={text} disabled={disabled}
            onChange={e => setText(e.target.value)}
            onKeyDown={e => {
                       if (e.key === 'Enter' && text) {
                           e.preventDefault();
                           send();
                       }
                   }}
            sx={{mt: 1}}
            slotProps={{
                       htmlInput: {'aria-label': 'Text to type on the device'},
                       input: {
                           endAdornment: <InputAdornment position="end">
                               <IconButton size="small" edge="end" aria-label="Type on the device" disabled={disabled || !text}
                                   onClick={send}><SendRounded fontSize="small"/></IconButton>
                           </InputAdornment>,
                       },
                   }}/>
    );
};

export default TypeField;
