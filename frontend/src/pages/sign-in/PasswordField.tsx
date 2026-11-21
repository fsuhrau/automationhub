import * as React from 'react';
import {IconButton, InputAdornment, TextField} from '@mui/material';
import {VisibilityOffRounded, VisibilityRounded} from '@mui/icons-material';

interface PasswordFieldProps {
    value: string,
    onChange: (value: string) => void,
    autoComplete: 'current-password' | 'new-password',
    error?: string,
    helperText?: string,
}

// a password input that can show what was typed
export const PasswordField: React.FC<PasswordFieldProps> = ({value, onChange, autoComplete, error, helperText}) => {
    const [shown, setShown] = React.useState(false);
    return <TextField label="Password" type={shown ? 'text' : 'password'} name="password" autoComplete={autoComplete} fullWidth={true}
        value={value} onChange={e => onChange(e.target.value)} error={!!error} helperText={error ?? helperText}
        slotProps={{
                          input: {
                              endAdornment: <InputAdornment position="end">
                                  <IconButton size="small" edge="end" aria-label={shown ? 'Hide the password' : 'Show the password'}
                                      onClick={() => setShown(v => !v)}>
                                      {shown ? <VisibilityOffRounded fontSize="small"/> : <VisibilityRounded fontSize="small"/>}
                                  </IconButton>
                              </InputAdornment>,
                          },
                      }}/>;
};
