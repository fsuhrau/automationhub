import Paper from "@mui/material/Paper";
import Grid from "@mui/material/Grid";
import Button from "@mui/material/Button";
import {
    Avatar,
    Chip,
    Divider,
    List,
    ListItem,
    ListItemAvatar,
    ListItemText,
    ListSubheader,
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
import {AppParameter, AppParameterOption, IAppData} from "../../types/app";
import {
    addAppParameter,
    createApp,
    getAllApps,
    removeAppParameter,
    updateApp,
    updateAppParameter
} from "../../services/app.service";
import {HubStateActions} from "../../application/HubState";
import EditAttributePopup, {EditAttribute} from "./EditAttributePopup";
import {useError} from "../../ErrorProvider";
import AppParameterPopup from "./AppParameterPopup";

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
                    <ListItem key={`app-liste-item-${app.id}`} onClick={() => {
                        onSelect(app.id)
                    }}>
                        <ListItemAvatar>
                            <Avatar variant={"rounded"}>
                                {icon}
                            </Avatar>
                        </ListItemAvatar>
                        <ListItemText primary={app.name} secondary={app.identifier}/>
                    </ListItem>
                ))}
            </List>)
    )
}


const AppsGroup: React.FC = () => {
    const {state, dispatch} = useHubState()
    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const [selectedAppID, setSelectedAppID] = useState<number | null>(state.apps === null ? null : state.apps.length === 0 ? null : state.apps[0].id);
    const selectedApp = state.apps?.find(a => a.id === selectedAppID);
    const iosApps = state.apps?.filter(a => a.platform === PlatformType.iOS);
    const androidApps = state.apps?.filter(a => a.platform === PlatformType.Android);
    const editorApps = state.apps?.filter(a => a.platform === PlatformType.Editor);
    const webApps = state.apps?.filter(a => a.platform === PlatformType.Web);
    const macApps = state.apps?.filter(a => a.platform === PlatformType.Mac);

    const selectApp = (id: number) => {
        setSelectedAppID(id);
    };

    const [uiState, setUiState] = useState<{
        showNewAppDialog: boolean,
        showAppParameterPopup: boolean,
        appParameter: AppParameter | null,
    }>({
        showNewAppDialog: false,
        showAppParameterPopup: false,
        appParameter: null,
    })

    const submitNewApp = (data: IAppData) => {
        createApp(projectIdentifier as string, data).then(app => {
            dispatch({type: HubStateActions.AppAdd, payload: app})
        }).catch(ex => setError(ex));
    }

    useEffect(() => {
        getAllApps(projectIdentifier as string).then(apps => {
            dispatch({
                type: HubStateActions.AppsUpdate,
                payload: apps,
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
        updateApp(projectIdentifier as string, selectedApp?.id as number, {
            ...selectedApp,
            [attribute]: value
        } as IAppData).then(app => {
            dispatch({
                type: HubStateActions.AppAttributeUpdate,
                payload: {
                    appId: selectedApp?.id,
                    attribute: attribute,
                    value: value
                }
            })
        }).catch(ex => setError(ex))
        onEditAttributeClose();
    };

    const handleEnvParameter = (param: AppParameter): void => {
        const index = selectedApp?.parameter?.findIndex(p => p.name === param.name);
        if (index === undefined || index < 0) {
            addAppParameter(projectIdentifier as string, selectedApp?.id as number, param).then(parameter => {
                    dispatch({
                        type: HubStateActions.AddAppParameter,
                        payload: parameter
                    })
                    setUiState(prevState => ({...prevState, showAppParameterPopup: false, appParameter: null}))
                }
            )
        } else {
            updateAppParameter(projectIdentifier as string, selectedApp?.id as number, param.id, param).then(parameter => {
                    dispatch({
                        type: HubStateActions.UpdateAppParameter,
                        payload: parameter
                    })
                    setUiState(prevState => ({...prevState, showAppParameterPopup: false, appParameter: null}))
                }
            )
        }
    }

    const handleDeleteParameter = (appId: number, parameterId: number) => {
        removeAppParameter(projectIdentifier as string, appId, parameterId).then(() => {
            dispatch({type: HubStateActions.DeleteAppParameter, payload: {appId: appId, id: parameterId}})
        })
    };


    return (<TitleCard title={"Apps"}>
        <EditAttributePopup attribute={changeAttributeDialogState.attribute} value={changeAttributeDialogState.value}
                            onSubmit={onEditAttributeSubmit} onClose={onEditAttributeClose}/>
        <AppParameterPopup open={uiState.showAppParameterPopup} parameter={uiState.appParameter}
                           onClose={() => setUiState(prevState => ({
                               ...prevState,
                               showAppParameterPopup: false,
                               appParameter: null
                           }))}
                           onSubmit={handleEnvParameter}/>
        <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
            <Grid container={true} spacing={1} padding={1}>
                <Grid size={{xs: 12, md: 12}} container={true} justifyContent={"flex-end"}>
                    <Button variant={"contained"}
                            onClick={() => setUiState(prevState => ({
                                ...prevState,
                                showNewAppDialog: true
                            }))}>{'Add App'}</Button>
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
                    <AppNavigation title={"unity Editor"} apps={editorApps}
                                   onSelect={selectApp} icon={<PlatformTypeIcon
                        platformType={PlatformType.Editor}/>}/>
                    <AppNavigation title={"Web apps"} apps={webApps} onSelect={selectApp}
                                   icon={<PlatformTypeIcon platformType={PlatformType.Web}/>}/>
                </Grid>
                <NewAppDialog open={uiState.showNewAppDialog} onSubmit={submitNewApp}
                              onClose={() => setUiState(prevState => ({...prevState, showNewAppDialog: false}))}/>
                <Grid size={{xs: 12, md: 1}}>
                    <Divider orientation={"vertical"}/>
                </Grid>
                <Grid container={true} size={{xs: 12, md: 8}} sx={{padding: 2}} alignContent={'flex-start'}>
                    {
                        selectedApp === null || selectedApp === undefined
                            ? (<Typography variant={"body1"} color={"dimgray"}>{'No App selected'}</Typography>)
                            : (<>
                                <Grid size={{xs: 12, md: 2}} container={true} alignItems={'center'}>
                                    <Typography variant={"caption"}>App ID</Typography>
                                </Grid>
                                <Grid size={{xs: 12, md: 10}} spacing={2} container={true} alignItems={'center'}>
                                    {selectedApp?.id}
                                </Grid>
                                <Grid size={12}><Divider/></Grid>
                                <Grid size={{xs: 12, md: 2}} container={true} alignItems={'center'}>
                                    <Typography variant={"caption"}>Bundle Identifier</Typography>
                                </Grid>
                                <Grid size={{xs: 12, md: 10}} spacing={2} container={true} alignItems={'center'}>
                                    <Typography variant={'body1'}>{selectedApp?.identifier}</Typography>
                                    {selectedApp?.identifier === "default_app" &&
                                        <IconButton aria-label="edit" size={'small'}
                                                    onClick={() => setChangeAttributeDialogState(prevState => ({
                                                        ...prevState,
                                                        attribute: 'identifier',
                                                        value: selectedApp!.identifier,
                                                    }))}><Edit/></IconButton>}
                                </Grid>
                                <Grid size={12}><Divider/></Grid>
                                <Grid size={{xs: 12, md: 2}} container={true} alignItems={'center'}>
                                    <Typography variant={"caption"}>App Name</Typography>
                                </Grid>
                                <Grid size={{xs: 12, md: 10}} spacing={2} container={true} alignItems={'center'}>
                                    <Typography variant={'body1'}>{selectedApp?.name}</Typography>
                                    <IconButton aria-label="edit" size={'small'}
                                                onClick={() => setChangeAttributeDialogState(prevState => ({
                                                    ...prevState,
                                                    attribute: 'name',
                                                    value: selectedApp!.name,
                                                }))}><Edit/></IconButton>
                                </Grid>
                                <Grid size={12}><Divider/></Grid>
                                <Grid size={{xs: 12, md: 2}} container={true}>
                                    <Typography variant={"caption"}>Parameter</Typography>
                                </Grid>
                                <Grid size={{xs: 12, md: 10}} spacing={1} container={true} alignItems={'center'}>
                                    {
                                        selectedApp?.parameter?.map(p => (
                                            <Grid container={true} size={12} key={`app_param_option_${p.id}`}>
                                                <Grid size={8}>
                                                    <Typography variant={"body2"}>
                                                        {'Name: '}{p.name}
                                                    </Typography>
                                                    <Typography variant={"body2"}>
                                                        {'Variant: '}{p.type.type}
                                                    </Typography>
                                                    <Typography variant={"body2"}>
                                                        {'Default Value: '}{p.type.defaultValue}
                                                    </Typography>
                                                    <Typography variant={"body2"}>
                                                        {'Options: '}
                                                    </Typography>
                                                    {p.type.type === 'option' && (p.type as AppParameterOption).options.map(o =>
                                                        <Chip key={`app_${p.id}_option_${o}`} label={o}/>)}
                                                </Grid>
                                                <Grid size={4} spacing={1} container={true}>
                                                    <Button variant={'contained'} onClick={() => setUiState(prevState => ({
                                                        ...prevState,
                                                        showAppParameterPopup: true,
                                                        appParameter: p
                                                    }))}>Edit</Button>
                                                    <Button variant={'contained'} color={"error"}
                                                            onClick={() => handleDeleteParameter(selectedApp.id, p.id)}>Delete</Button>
                                                </Grid>
                                                <Grid size={12}><Divider/></Grid>
                                            </Grid>
                                        ))
                                    }
                                    <Button variant={'contained'} onClick={() => setUiState(prevState => ({
                                        ...prevState,
                                        showAppParameterPopup: true,
                                        appParameter: null
                                    }))}>Add new Parameter</Button>
                                </Grid>
                            </>)
                    }
                </Grid>
            </Grid>
        </Paper>
    </TitleCard>)
}

export default AppsGroup;