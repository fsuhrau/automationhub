import React from 'react';
import {ToggleButton, ToggleButtonGroup} from '@mui/material';
import {useColorScheme} from '@mui/material/styles';
import {Section} from '../../../components/common';

// AppearanceSection picks light, dark or the system's color mode
const AppearanceSection: React.FC = () => {
    const {mode, setMode} = useColorScheme();
    return <Section title="Appearance" description="Light, dark, or as your system is set. Stored in this browser.">
        {mode && <ToggleButtonGroup size="small" exclusive={true} value={mode} onChange={(_, v) => v && setMode(v)} aria-label="Color mode">
            <ToggleButton value="system">System</ToggleButton>
            <ToggleButton value="light">Light</ToggleButton>
            <ToggleButton value="dark">Dark</ToggleButton>
        </ToggleButtonGroup>}
    </Section>;
};

export default AppearanceSection;
