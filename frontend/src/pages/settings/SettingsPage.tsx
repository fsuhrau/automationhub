import React, {useState} from 'react';
import Grid from "@mui/material/Grid";
import {Box, Divider, Tab, Tabs,} from '@mui/material';
import {TitleCard} from "../../components/title.card.component";
import AccessTokenTab from "./AccessTokenTab";
import NodesTab from "./NodesTab";
import AppsGroup from "./AppsGroup";
import ProjectGroup from "./ProjectGroup";
import UsersTab from "./UsersTab";

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
}

const TabPanel: React.FC<TabPanelProps> = (props: TabPanelProps) => {
    const {children, value, index, ...other} = props;
    return (
        <Grid

            size={{xs: 12, md: 12}}
            style={{width: '100%'}}
            role="tabpanel"
            hidden={value !== index}
            id={`simple-tabpanel-${index}`}
            aria-labelledby={`simple-tab-${index}`}
            {...other}
        >
            {value === index && children}
        </Grid>
    )
};

const SettingsPage: React.FC = () => {

    function a11yProps(index: number): Map<string, string> {
        return new Map([
            ['id', `simple-tab-${index}`],
            ['aria-controls', `simple-tabpanel-${index}`],
        ]);
    }

    const [tabIndex, setTabIndex] = useState(0);
    const onTabIndexChange = (event: React.ChangeEvent<{}>, newValue: number): void => {
        setTabIndex(newValue);
    };

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title={'Project Settings'}>
                <Grid container={true} spacing={2}
                      sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}
                >
                    <Grid size={{xs: 12, md: 12}}>
                        <Tabs
                            value={tabIndex}
                            onChange={onTabIndexChange}
                            indicatorColor="primary"
                            textColor="inherit"
                            aria-label="secondary tabs"
                        >
                            <Tab label="General" {...a11yProps(0)} />
                            <Tab label="Access Tokens" {...a11yProps(1)} />
                            <Tab label="Users and Permissions" {...a11yProps(2)} />
                            <Tab label="Notifications" {...a11yProps(3)} />
                            <Tab label="Nodes" {...a11yProps(4)} />
                        </Tabs>
                        <Divider/>
                    </Grid>
                    <Grid container={true} size={{xs: 12, md: 12}} alignItems={"center"} justifyContent={"center"}
                          sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}
                    >
                        { /* General */}
                        <TabPanel index={tabIndex} value={0}>
                            <ProjectGroup/>
                            <AppsGroup/>
                        </TabPanel>
                        { /* Access Tokens */}
                        <TabPanel index={tabIndex} value={1}>
                            <AccessTokenTab/>
                        </TabPanel>
                        { /* User and Permissions */}
                        <TabPanel index={tabIndex} value={2}>
                            <UsersTab />
                        </TabPanel>
                        { /* Notifications */}
                        <TabPanel index={tabIndex} value={3}>
                            <TitleCard title={"Notifications"}>
                                not implemented
                            </TitleCard>
                        </TabPanel>
                        { /* Nodes */}
                        <TabPanel index={tabIndex} value={4}>
                            <NodesTab/>
                        </TabPanel>
                    </Grid>
                </Grid>
            </TitleCard>
        </Box>);
};

export default SettingsPage;
