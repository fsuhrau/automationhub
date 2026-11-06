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
import {useLocation, useNavigate, useParams} from "react-router-dom";
import DevicesRoundedIcon from "@mui/icons-material/DevicesRounded";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useHubState} from "../hooks/HubStateProvider";
import { redirect } from "react-router";

export default function MenuContent() {

    const {state} = useHubState()

    const {projectIdentifier} = useProjectContext();
    const {appId: urlAppId} = useParams<{ appId: string }>();
    const navigate = useNavigate();
    const location = useLocation();

    const stAppId = localStorage.getItem('appId')
    let appId = "";
    if (urlAppId) {
        appId = urlAppId.replace("app:", "")
    } else if (stAppId) {
        appId = stAppId
    }

    const mainListItems = [
        // app scoped pages only exist once an app is selected
        { text: 'Home', ref: appId ? `/project/${projectIdentifier}/app:${appId}/home` : `/project/${projectIdentifier}`, icon: <HomeRoundedIcon />, disabled: false },
        { text: 'Tests', ref: `/project/${projectIdentifier}/app:${appId}/tests`, icon: <AnalyticsRoundedIcon />, disabled: !appId },
        { text: 'Apps / Bundles', ref: `/project/${projectIdentifier}/app:${appId}/bundles`, icon: <PeopleRoundedIcon />, disabled: !appId },
        { text: 'Device Pool', ref: `/project/${projectIdentifier}/devices`, icon: <DevicesRoundedIcon />, disabled: false },
    ];

    const secondaryListItems = [
        { text: 'Settings', external: false, ref: `/project/${projectIdentifier}/settings`, icon: <SettingsRoundedIcon /> },
        { text: 'About', external: true, ref: `https://github.com/fsuhrau/automationhub`, icon: <InfoRoundedIcon /> },
        /* { text: 'Feedback', external: false, ref: '', icon: <HelpRoundedIcon /> },*/
    ];

  return (
    <Stack sx={{ flexGrow: 1, p: 1, justifyContent: 'space-between' }}>
      <List dense>
        {mainListItems.map((item, index) => (
          <ListItem key={index} disablePadding sx={{ display: 'block' }}>
            <ListItemButton disabled={item.disabled} selected={location.pathname.startsWith(item.ref)} onClick={event => {
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
                if (item.external) {
                    window.open(item.ref, '_blank');
                } else {
                    navigate(item.ref);
                }
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
