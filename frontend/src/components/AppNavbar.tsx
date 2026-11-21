import * as React from 'react';
import {styled} from '@mui/material/styles';
import AppBar from '@mui/material/AppBar';
import Stack from '@mui/material/Stack';
import MuiToolbar from '@mui/material/Toolbar';
import {tabsClasses} from '@mui/material/Tabs';
import Typography from '@mui/material/Typography';
import MenuRoundedIcon from '@mui/icons-material/MenuRounded';
import SideMenuMobile from './SideMenuMobile';
import MenuButton from './MenuButton';
import ColorModeIconDropdown from '../shared-theme/ColorModeIconDropdown';
import HubLogo from './HubLogo';

const Toolbar = styled(MuiToolbar)({
    width: '100%',
    padding: '12px',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'start',
    justifyContent: 'center',
    gap: '12px',
    flexShrink: 0,
    [`& ${tabsClasses.flexContainer}`]: {
        gap: '8px',
        p: '8px',
        pb: 0,
    },
});

const barSx = {
    display: {xs: 'auto', md: 'none'},
    boxShadow: 0,
    bgcolor: 'background.paper',
    backgroundImage: 'none',
    borderBottom: '1px solid',
    borderColor: 'divider',
    top: 'var(--template-frame-height, 0px)',
};

const Brand: React.FC = () => (
    <Stack direction="row" spacing={1} sx={{justifyContent: 'center', mr: 'auto'}}>
        <HubLogo sx={{fontSize: 24}}/>
        <Typography variant="h4" component="div" sx={{color: 'text.primary'}}>Automation Hub</Typography>
    </Stack>
);

// AppNavbar is the top bar of small screens; it opens the menu as a drawer
export default function AppNavbar(): React.ReactElement {
    const [open, setOpen] = React.useState(false);
    const toggleDrawer = (newOpen: boolean) => (): void => setOpen(newOpen);

    return (
        <AppBar position="fixed" sx={barSx}>
            <Toolbar variant="regular">
                <Stack direction="row" sx={{alignItems: 'center', flexGrow: 1, width: '100%', gap: 1}}>
                    <Brand/>
                    <ColorModeIconDropdown/>
                    <MenuButton aria-label="Open navigation" onClick={toggleDrawer(true)}>
                        <MenuRoundedIcon/>
                    </MenuButton>
                    <SideMenuMobile open={open} toggleDrawer={toggleDrawer}/>
                </Stack>
            </Toolbar>
        </AppBar>
    );
}
