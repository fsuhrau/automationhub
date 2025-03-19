import React, {useEffect, useState} from 'react';
import Paper from '@mui/material/Paper';
import Button from '@mui/material/Button';
import Grid from "@mui/material/Grid2";
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import Moment from 'react-moment';
import {
    Avatar,
    Box,
    Dialog,
    DialogActions,
    DialogContent,
    Divider,
    List,
    ListItem,
    ListItemAvatar,
    ListItemText,
    ListSubheader,
    Tab,
    Tabs,
    TextField,
    Typography
} from '@mui/material';
import {
    createAccessToken,
    deleteAccessToken,
    getAccessTokens,
    NewAccessTokenRequest,
} from '../../services/settings.service';
import IAccessTokenData from '../../types/access.token';
import CopyToClipboard from '../../components/copy.clipboard.component';
import {ContentCopy, Edit} from '@mui/icons-material';
import IconButton from '@mui/material/IconButton';
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';
import {AdapterDayjs} from '@mui/x-date-pickers/AdapterDayjs';
import {LocalizationProvider, MobileDatePicker} from '@mui/x-date-pickers';
import {Dayjs} from 'dayjs';
import {PlatformType} from "../../types/platform.type.enum";
import {IAppData} from "../../types/app";
import {TitleCard} from "../../components/title.card.component";
import {updateProject} from "../../project/project.service";
import {HubStateActions} from "../../application/HubState";
import IProject from "../../project/project";
import {createApp, updateApp} from "../../services/app.service";
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useHubState} from "../../hooks/HubStateProvider";
import PlatformTypeIcon from "../../components/PlatformTypeIcon";
import NewAppDialog from "../apps/newapp.dialog";

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

interface AppNavigationProps {
    title: string,
    apps: IAppData[] | undefined,
    onSelect: (id: number) => void,
    icon: React.ReactNode,
}

const AppNavigation: React.FC<AppNavigationProps> = (props: AppNavigationProps) => {

    const {title, apps, onSelect, icon} = props;

    return (apps === undefined || apps.length === 0 ? null : (
            <List sx={{width: '100%'}} subheader={<ListSubheader>{title}</ListSubheader>}>
                {apps.map(app => (
                    <ListItem key={`app-liste-item-${app.ID}`} onClick={() => {
                        onSelect(app.ID)
                    }}>
                        <ListItemAvatar>
                            <Avatar variant={"rounded"}>
                                {icon}
                            </Avatar>
                        </ListItemAvatar>
                        <ListItemText primary={app.Name} secondary={app.Identifier}/>
                    </ListItem>
                ))}
            </List>)
    )
}

const SettingsPage: React.FC = () => {

    const {state, dispatch} = useHubState()
    const {project, projectIdentifier} = useProjectContext();

    function a11yProps(index: number): Map<string, string> {
        return new Map([
            ['id', `simple-tab-${index}`],
            ['aria-controls', `simple-tabpanel-${index}`],
        ]);
    }

    const [value, setValue] = useState(0);
    const handleChange = (event: React.ChangeEvent<{}>, newValue: number): void => {
        setValue(newValue);
    };

    const [accessTokens, setAccessTokens] = useState<IAccessTokenData[]>([]);

    useEffect(() => {
        getAccessTokens(projectIdentifier).then(response => {
            setAccessTokens(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, [projectIdentifier]);

    const handleDeleteAccessToken = (accessTokenID: number): void => {
        deleteAccessToken(projectIdentifier, accessTokenID).then(value => {
            setAccessTokens(prevState => {
                const newState = [...prevState];
                const index = newState.findIndex(value1 => value1.ID == accessTokenID);
                if (index > -1) {
                    newState.splice(index, 1);
                }
                return newState;
            });
        });
    };

    const [newToken, setNewToken] = useState<NewAccessTokenRequest>({
        Name: '',
        ExpiresAt: null,
    });

    const createNewAccessToken = (): void => {
        createAccessToken(projectIdentifier, newToken).then(response => {
            setAccessTokens(prevState => {
                const newState = [...prevState];
                newState.push(response.data);
                return newState;
            });
            setNewToken({
                Name: '',
                ExpiresAt: null,
            })
        }).catch(ex => {
            console.log(ex);
        });
    };

    const [selectedAppID, setSelectedAppID] = useState<number | null>(project.Apps === undefined ? null : project.Apps.length === 0 ? null : project.Apps[0].ID);
    const selectedApp = project.Apps.find(a => a.ID === selectedAppID);

    const iosApps = project.Apps.filter(a => a.Platform === PlatformType.iOS);
    const androidApps = project.Apps.filter(a => a.Platform === PlatformType.Android);
    const editorApps = project.Apps.filter(a => a.Platform === PlatformType.Editor);
    const webApps = project.Apps.filter(a => a.Platform === PlatformType.Web);
    const macApps = project.Apps.filter(a => a.Platform === PlatformType.Mac);

    const selectApp = (id: number) => {
        setSelectedAppID(id);
    };

    type ChangeAttributeState = {
        context?: 'project' | 'app',
        attribute: string,
        value: string,
        open: boolean,
    }

    const [changeAttributeDialogState, setChangeAttributeDialogState] = useState<ChangeAttributeState>({
        attribute: '',
        value: '',
        open: false
    });

    const handleEditClose = () => {
        setChangeAttributeDialogState(prevState => ({...prevState, attribute: '', value: '', open: false}))
    };

    const handleEditSubmit = () => {
        if (changeAttributeDialogState.context === 'project') {
            updateProject(projectIdentifier as string, {
                ...project,
                [changeAttributeDialogState.attribute]: changeAttributeDialogState.value
            } as IProject).then(response => {
                dispatch({
                    type: HubStateActions.UpdateProjectAttribute,
                    payload: {
                        projectIdentifier: projectIdentifier,
                        attribute: changeAttributeDialogState.attribute,
                        value: changeAttributeDialogState.value
                    }
                })
            })
        } else if (changeAttributeDialogState.context === 'app') {
            updateApp(projectIdentifier as string, selectedApp?.ID as number, {
                ...selectedApp,
                [changeAttributeDialogState.attribute]: changeAttributeDialogState.value
            } as IAppData).then(response => {
                dispatch({
                    type: HubStateActions.UpdateAppAttribute,
                    payload: {
                        projectIdentifier: projectIdentifier,
                        appId: selectedApp?.ID,
                        attribute: changeAttributeDialogState.attribute,
                        value: changeAttributeDialogState.value
                    }
                })
            })
        }

        handleEditClose();
    };

    const [showNewAppDialog, setShowNewAppDialog] = useState<boolean>(false);

    const submitNewApp = (data: IAppData) => {
        createApp(projectIdentifier as string, data).then(response => {
            dispatch({type: HubStateActions.AddNewApp, payload: response.data})
        }).catch(ex => console.log(ex));
    }

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <Dialog open={changeAttributeDialogState.open} onClose={handleEditClose}>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        id="change_attribute"
                        placeholder={changeAttributeDialogState.attribute}
                        value={changeAttributeDialogState.value}
                        onChange={(e) => setChangeAttributeDialogState(prevState => ({
                            ...prevState,
                            value: e.target.value
                        }))}
                        fullWidth
                        variant="standard"
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleEditClose}>Cancel</Button>
                    <Button onClick={handleEditSubmit}>Save</Button>
                </DialogActions>
            </Dialog>
            <TitleCard title={'Project Settings'}>
                <Grid container={true} spacing={2}
                      sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}
                >
                    <Grid size={{xs: 12, md: 12}}>
                        <Tabs
                            value={value}
                            onChange={handleChange}
                            indicatorColor="primary"
                            textColor="inherit"
                            aria-label="secondary tabs"
                        >
                            <Tab label="General" {...a11yProps(0)} />
                            <Tab label="Access Tokens" {...a11yProps(1)} />
                            <Tab label="Users and Permissions" {...a11yProps(2)} />
                            <Tab label="Notifications" {...a11yProps(3)} />
                        </Tabs>
                        <Divider/>
                    </Grid>
                    <Grid container={true} size={{xs: 12, md: 12}} alignItems={"center"} justifyContent={"center"}
                          sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}
                    >
                        { /* General */}
                        <TabPanel index={value} value={0}>
                            <TitleCard title={"Project"}>
                                <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
                                    <Box sx={{padding: 2}}>
                                        <Grid container={true} spacing={4}>
                                            <Grid size={{xs: 12, md: 1}}>
                                                <Typography variant={"body1"} color={"dimgray"}>Project
                                                    name</Typography>
                                            </Grid>
                                            <Grid size={{xs: 12, md: 11}} container={true} spacing={2}>
                                                {project.Name}<IconButton aria-label="edit" size={'small'}
                                                                          onClick={() => setChangeAttributeDialogState(prevState => ({
                                                                              ...prevState,
                                                                              context: "project",
                                                                              attribute: 'Name',
                                                                              open: true
                                                                          }))}><Edit/></IconButton>
                                            </Grid>
                                            <Grid size={{xs: 12, md: 1}}>
                                                <Typography variant={"body1"} color={"dimgray"}>Project ID</Typography>
                                            </Grid>
                                            <Grid size={{xs: 12, md: 11}} container={true} spacing={2}>
                                                {projectIdentifier}{projectIdentifier === "default_project" &&
                                                <IconButton aria-label="edit" size={'small'}
                                                            onClick={() => setChangeAttributeDialogState(prevState => ({
                                                                ...prevState,
                                                                context: "project",
                                                                attribute: 'Identifier',
                                                                open: true
                                                            }))}><Edit/></IconButton>}
                                            </Grid>
                                        </Grid>
                                    </Box>
                                </Paper>
                            </TitleCard> { /* Notifications */}
                            <TabPanel index={value} value={3}>
                                <TitleCard title={"Notifications"}>
                                    not implemented
                                </TitleCard>
                            </TabPanel>

                            <TitleCard title={"Apps"}>
                                <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
                                    <Grid container={true} spacing={1} padding={1}>
                                        <Grid size={{xs: 12, md: 12}} container={true} justifyContent={"flex-end"}>
                                            <Button variant={"contained"} onClick={() => setShowNewAppDialog(true)}>Add App</Button>
                                        </Grid>
                                        <Grid size={12}>
                                            <Divider />
                                        </Grid>
                                        <Grid container={true} size={{xs: 12, md: 2}}>
                                            <AppNavigation title={"Android apps"} apps={androidApps}
                                                           onSelect={selectApp} icon={<PlatformTypeIcon
                                                platformType={PlatformType.Android}/>}/>
                                            <AppNavigation title={"Apple apps"} apps={iosApps} onSelect={selectApp}
                                                           icon={<PlatformTypeIcon platformType={PlatformType.iOS}/>}/>
                                            <AppNavigation title={"MacOS apps"} apps={macApps} onSelect={selectApp}
                                                           icon={<PlatformTypeIcon platformType={PlatformType.Mac}/>}/>
                                            <AppNavigation title={"Unity Editor"} apps={editorApps}
                                                           onSelect={selectApp} icon={<PlatformTypeIcon
                                                platformType={PlatformType.Editor}/>}/>
                                            <AppNavigation title={"Web apps"} apps={webApps} onSelect={selectApp}
                                                           icon={<PlatformTypeIcon platformType={PlatformType.Web}/>}/>
                                        </Grid>
                                        <NewAppDialog open={showNewAppDialog} onSubmit={submitNewApp}
                                                      onClose={() => setShowNewAppDialog(false)}/>
                                        <Grid size={{xs: 12, md: 1}}>
                                            <Divider orientation={"vertical"}/>
                                        </Grid>
                                        <Grid container={true} size={{xs: 12, md: 8}} sx={{padding: 2}}>
                                            {
                                                selectedApp === null ? (<Typography variant={"body1"} color={"dimgray"}>No
                                                    Apps</Typography>) : (<>
                                                    <Grid size={{xs: 12, md: 2}}>
                                                        <Typography variant={"caption"}>App ID</Typography>
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 10}} container={true} spacing={2}>
                                                        {selectedApp?.ID}
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 2}}>
                                                        <Typography variant={"caption"}>App Name</Typography>
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 10}} container={true} spacing={2}>
                                                        {selectedApp?.Name}<IconButton aria-label="edit" size={'small'}
                                                                                       onClick={() => setChangeAttributeDialogState(prevState => ({
                                                                                           ...prevState,
                                                                                           context: "app",
                                                                                           attribute: 'Name',
                                                                                           open: true
                                                                                       }))}><Edit/></IconButton>
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 2}}>
                                                        <Typography variant={"caption"}>Bundle Identifier</Typography>
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 10}} container={true} spacing={2}>
                                                        {selectedApp?.Identifier}{selectedApp?.Identifier === "default_app" &&
                                                        <IconButton aria-label="edit" size={'small'}
                                                                    onClick={() => setChangeAttributeDialogState(prevState => ({
                                                                        ...prevState,
                                                                        context: "app",
                                                                        attribute: 'Identifier',
                                                                        open: true
                                                                    }))}><Edit/></IconButton>}
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 2}}>
                                                        <Typography variant={"caption"}>Default Parameter</Typography>
                                                    </Grid>
                                                    <Grid size={{xs: 12, md: 10}} container={true} spacing={2}>
                                                        {selectedApp?.DefaultParameter}<IconButton aria-label="edit"
                                                                                                   size={'small'}
                                                                                                   onClick={() => setChangeAttributeDialogState(prevState => ({
                                                                                                       ...prevState,
                                                                                                       context: "app",
                                                                                                       attribute: 'DefaultParameter',
                                                                                                       open: true
                                                                                                   }))}><Edit/></IconButton>
                                                    </Grid>
                                                </>)
                                            }
                                        </Grid>
                                    </Grid>
                                </Paper>
                            </TitleCard>
                        </TabPanel>
                        { /* Access Tokens */}
                        <TabPanel index={value} value={1}>
                            <TitleCard title={"Access Tokens"}>
                                <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
                                    <Box sx={{width: '100%'}}>
                                        <Box sx={{p: 3}}>
                                            <Table size="small" aria-label="a dense table">
                                                <TableHead>
                                                    <TableRow>
                                                        <TableCell>Name</TableCell>
                                                        <TableCell align="right">Token</TableCell>
                                                        <TableCell align="right">Expires</TableCell>
                                                        <TableCell></TableCell>
                                                    </TableRow>
                                                </TableHead>
                                                <TableBody>
                                                    {accessTokens.map((accessToken) => <TableRow key={accessToken.ID}>
                                                        <TableCell>{accessToken.Name}</TableCell>
                                                        <TableCell>
                                                            <Grid container={true} direction={"row"} spacing={1} justifyContent={"right"} alignItems={"center"}>
                                                                <Grid>
                                                                    {accessToken.Token}
                                                                </Grid>
                                                                <Grid>
                                                                    <CopyToClipboard>
                                                                        {({copy}) => (
                                                                            <IconButton
                                                                                color={'primary'}
                                                                                size={'small'}
                                                                                onClick={() => copy(accessToken.Token)}
                                                                            >
                                                                                <ContentCopy/>
                                                                            </IconButton>
                                                                        )}
                                                                    </CopyToClipboard>

                                                                </Grid>
                                                            </Grid>
                                                        </TableCell>
                                                        <TableCell align="right">{accessToken.ExpiresAt !== null ? (
                                                            <Moment
                                                                format="YYYY/MM/DD HH:mm:ss">{accessToken.ExpiresAt}</Moment>) : ('Unlimited')}</TableCell>
                                                        <TableCell>
                                                            <IconButton color="secondary" size="small" onClick={() => {
                                                                handleDeleteAccessToken(accessToken.ID as number);
                                                            }}>
                                                                <DeleteForeverIcon/>
                                                            </IconButton>
                                                        </TableCell>
                                                    </TableRow>)}

                                                    <TableRow>
                                                        <TableCell>
                                                            <TextField id="new_name" placeholder="Name"
                                                                       variant="outlined"
                                                                       fullWidth={true}
                                                                       value={newToken.Name} size="small"
                                                                       onChange={event => setNewToken(prevState => ({
                                                                           ...prevState,
                                                                           Name: event.target.value
                                                                       }))}/>
                                                        </TableCell>
                                                        <TableCell></TableCell>
                                                        <TableCell align="right">
                                                            <LocalizationProvider dateAdapter={AdapterDayjs}>
                                                                <MobileDatePicker
                                                                    value={newToken.ExpiresAt}
                                                                    onChange={(newValue: Dayjs | null) => {
                                                                        setNewToken(prevState => ({
                                                                            ...prevState,
                                                                            ExpiresAt: newValue
                                                                        }))
                                                                    }}
                                                                />
                                                            </LocalizationProvider>
                                                        </TableCell>
                                                        <TableCell>
                                                            <Button variant="contained" color="primary" size="small"
                                                                    onClick={createNewAccessToken}>
                                                                Add
                                                            </Button>
                                                        </TableCell>
                                                    </TableRow>
                                                </TableBody>
                                            </Table>
                                        </Box>
                                    </Box>
                                </Paper>
                            </TitleCard>
                        </TabPanel>
                        { /* User and Permissions */}
                        <TabPanel index={value} value={2}>
                            <TitleCard title={"Users and Permission"}>
                                not implemented
                            </TitleCard>
                        </TabPanel>
                        { /* Notifications */}
                        <TabPanel index={value} value={3}>
                            <TitleCard title={"Notifications"}>
                                not implemented
                            </TitleCard>
                        </TabPanel>
                    </Grid>
                </Grid>
            </TitleCard>
        </Box>);
};

export default SettingsPage;
