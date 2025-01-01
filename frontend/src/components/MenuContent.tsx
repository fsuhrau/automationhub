import * as React from 'react';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Stack from '@mui/material/Stack';
import HomeRoundedIcon from '@mui/icons-material/HomeRounded';
import AnalyticsRoundedIcon from '@mui/icons-material/AnalyticsRounded';
import PeopleRoundedIcon from '@mui/icons-material/PeopleRounded';
import SettingsRoundedIcon from '@mui/icons-material/SettingsRounded';
import InfoRoundedIcon from '@mui/icons-material/InfoRounded';
import HelpRoundedIcon from '@mui/icons-material/HelpRounded';
import {useNavigate} from "react-router-dom";
import DevicesRoundedIcon from "@mui/icons-material/DevicesRounded";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";

export default function MenuContent() {

    const {state} = useHubState()

    const {projectIdentifier} = useProjectContext();

    const navigate = useNavigate();

    const mainListItems = [
        { text: 'Home', ref: `/project/${projectIdentifier}`, icon: <HomeRoundedIcon /> },
        { text: 'Tests', ref: state.appId ? `/project/${projectIdentifier}/app/tests` : `/project/${projectIdentifier}/tests`, icon: <AnalyticsRoundedIcon /> },
        { text: 'Apps', ref: state.appId ? `/project/${projectIdentifier}/app/bundles` : `/project/${projectIdentifier}/bundles`, icon: <PeopleRoundedIcon /> },
        { text: 'Devices', ref: `/project/${projectIdentifier}/devices`, icon: <DevicesRoundedIcon /> },
    ];

    const secondaryListItems = [
        { text: 'Settings', ref: `/project/${projectIdentifier}/settings`, icon: <SettingsRoundedIcon /> },
        { text: 'About', ref: `https://github.com/fsuhrau/automationhub`, icon: <InfoRoundedIcon /> },
        { text: 'Feedback', ref: '', icon: <HelpRoundedIcon /> },
    ];

  return (
    <Stack sx={{ flexGrow: 1, p: 1, justifyContent: 'space-between' }}>
      <List dense>
        {mainListItems.map((item, index) => (
          <ListItem key={index} disablePadding sx={{ display: 'block' }}>
            <ListItemButton selected={index === 0} onClick={event => {
                navigate(item.ref)
            }}>
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>

      <List dense>
        {secondaryListItems.map((item, index) => (
          <ListItem key={index} disablePadding sx={{ display: 'block' }}>
            <ListItemButton onClick={event => {
                navigate(item.ref)
            }}>
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Stack>
  );
}
