import * as React from 'react';
import {Box, Card, Stack, Typography} from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import type {SystemStyleObject, Theme} from '@mui/system';
import HubLogo from '../../components/HubLogo';
import ColorModeSelect from '../../shared-theme/ColorModeSelect';

interface AuthPageProps {
    title: string,
    subtitle?: React.ReactNode,
    children: React.ReactNode,
}

const backdropSx = (theme: Theme): SystemStyleObject<Theme> => ({
    minHeight: '100dvh', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 2,
    backgroundImage: 'radial-gradient(ellipse at 50% 30%, hsl(210, 100%, 97%), hsl(0, 0%, 100%))',
    ...theme.applyStyles('dark', {backgroundImage: 'radial-gradient(at 50% 30%, hsla(210, 100%, 16%, 0.5), hsl(220, 30%, 5%))'}),
});

// AuthPage is the frame of the sign in and sign up pages: the product, a
// card in the middle, the color mode in the corner
export const AuthPage: React.FC<AuthPageProps> = ({title, subtitle, children}) =>
    <>
        <CssBaseline enableColorScheme={true}/>
        <ColorModeSelect size="small" sx={{position: 'fixed', top: '1rem', right: '1rem'}} inputProps={{'aria-label': 'Color mode'}}/>
        <Box sx={backdropSx}>
            <Card variant="outlined" sx={{width: '100%', maxWidth: 420, p: {xs: 3, sm: 4}}}>
                <Stack spacing={1} sx={{mb: 3}}>
                    <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 1}}>
                        <HubLogo sx={{fontSize: 28}}/>
                        <Typography variant="subtitle1" sx={{fontWeight: 600}}>Automation Hub</Typography>
                    </Stack>
                    <Typography component="h1" variant="h5">{title}</Typography>
                    {subtitle && <Typography variant="body2" color="text.secondary">{subtitle}</Typography>}
                </Stack>
                {children}
            </Card>
        </Box>
    </>;
