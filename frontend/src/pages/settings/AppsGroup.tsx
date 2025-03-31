import Paper from "@mui/material/Paper";
import Grid from "@mui/material/Grid2";
import Button from "@mui/material/Button";
import {
    Avatar,
    Dialog,
    DialogActions,
    DialogContent,
    Divider,
    List,
    ListItem, ListItemAvatar, ListItemText,
    ListSubheader,
    TextField,
    Typography
} from "@mui/material";
import PlatformTypeIcon from "../../components/PlatformTypeIcon";
import {PlatformType} from "../../types/platform.type.enum";
import NewAppDialog from "../apps/newapp.dialog";
import IconButton from "@mui/material/IconButton";
import {Edit} from "@mui/icons-material";
import {TitleCard} from "../../components/title.card.component";
import React, {useEffect, useState} from "react";
import {useHubState} from "../../hooks/HubStateProvider";
import {useProjectContext} from "../../hooks/ProjectProvider";
import {IAppData} from "../../types/app";
import {createApp, getAllApps, updateApp} from "../../services/app.service";
import {HubStateActions} from "../../application/HubState";
import {updateProject} from "../../project/project.service";
import IProject from "../../project/project";
import EditAttributePopup, {EditAttribute} from "./EditAttributePopup";
import {useError} from "../../ErrorProvider";

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


const AppsGroup: React.FC = () => {
    const {state, dispatch} = useHubState()
    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const [selectedAppID, setSelectedAppID] = useState<number | null>(state.apps === null ? null : state.apps.length === 0 ? null : state.apps[0].ID);
    const selectedApp = state.apps?.find(a => a.ID === selectedAppID);
    const iosApps = state.apps?.filter(a => a.Platform === PlatformType.iOS);
    const androidApps = state.apps?.filter(a => a.Platform === PlatformType.Android);
    const editorApps = state.apps?.filter(a => a.Platform === PlatformType.Editor);
    const webApps = state.apps?.filter(a => a.Platform === PlatformType.Web);
    const macApps = state.apps?.filter(a => a.Platform === PlatformType.Mac);

    const selectApp = (id: number) => {
        setSelectedAppID(id);
    };

    const [showNewAppDialog, setShowNewAppDialog] = useState<boolean>(false);
    const submitNewApp = (data: IAppData) => {
        createApp(projectIdentifier as string, data).then(response => {
            dispatch({type: HubStateActions.AppAdd, payload: response.data})
        }).catch(ex => setError(ex));
    }

    useEffect(() => {
        getAllApps(projectIdentifier as string).then(response => {
            dispatch({
                type: HubStateActions.AppsUpdate,
                payload: response.data,
            })
        }).catch(ex => {
            setError(ex)
        });
    }, [projectIdentifier]);

    const [changeAttributeDialogState, setChangeAttributeDialogState] = useState<EditAttribute>({
        attribute: '',
        value: '',
    });

    const onEditAttributeClose = () => {
        setChangeAttributeDialogState(prevState => ({...prevState, attribute: '', value: ''}))
    };

    const onEditAttributeSubmit = (attribute: string, value: string) => {
        updateApp(projectIdentifier as string, selectedApp?.ID as number, {
            ...selectedApp,
            [attribute]: value
        } as IAppData).then(response => {
            dispatch({
                type: HubStateActions.AppAttributeUpdate,
                payload: {
                    appId: selectedApp?.ID,
                    attribute: attribute,
                    value: value
                }
            })
        }).catch(ex => setError(ex))
        onEditAttributeClose();
    };


    return (<TitleCard title={"Apps"}>
        <EditAttributePopup attribute={changeAttributeDialogState.attribute} value={changeAttributeDialogState.value} onSubmit={onEditAttributeSubmit} onClose={onEditAttributeClose} />
        <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
            <Grid container={true} spacing={1} padding={1}>
                <Grid size={{xs: 12, md: 12}} container={true} justifyContent={"flex-end"}>
                    <Button variant={"contained"} onClick={() => setShowNewAppDialog(true)}>Add App</Button>
                </Grid>
                <Grid size={12}>
                    <Divider/>
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
                                                                   attribute: 'Name',
                                                                   value: selectedApp!.Name,
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
                                                attribute: 'Identifier',
                                                value: selectedApp!.Identifier,
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
                                                                               attribute: 'DefaultParameter',
                                                                               value: selectedApp!.DefaultParameter,
                                                                           }))}><Edit/></IconButton>
                            </Grid>
                        </>)
                    }
                </Grid>
            </Grid>
        </Paper>
    </TitleCard>)
}

export default AppsGroup;