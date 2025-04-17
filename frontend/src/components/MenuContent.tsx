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
        { text: 'Home', ref: `/project/${projectIdentifier}/app:${appId}/home`, icon: <HomeRoundedIcon /> },
        { text: 'Tests', ref: appId ? `/project/${projectIdentifier}/app:${appId}/tests` : `/project/${projectIdentifier}/tests`, icon: <AnalyticsRoundedIcon /> },
        { text: 'Apps', ref: appId ? `/project/${projectIdentifier}/app:${appId}/bundles` : `/project/${projectIdentifier}/bundles`, icon: <PeopleRoundedIcon /> },
        { text: 'Devices', ref: `/project/${projectIdentifier}/devices`, icon: <DevicesRoundedIcon /> },
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
            <ListItemButton selected={location.pathname.startsWith(item.ref)} onClick={event => {
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
