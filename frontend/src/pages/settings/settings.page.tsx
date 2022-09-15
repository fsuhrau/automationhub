import React, { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import { useNavigate, useParams } from 'react-router-dom';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import Moment from 'react-moment';
import {
    Avatar,
    Box,
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
import { Android, Apple, ContentCopy, Web } from '@mui/icons-material';
import IconButton from '@mui/material/IconButton';
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { LocalizationProvider, MobileDatePicker } from '@mui/x-date-pickers';
import { Dayjs } from 'dayjs';
import { ApplicationProps } from "../../application/application.props";
import { PlatformType } from "../../types/platform.type.enum";
import { IAppData } from "../../types/app";
import { TitleCard } from "../../components/title.card.component";

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
}

const TabPanel: React.FC<TabPanelProps> = (props: TabPanelProps) => {
    const {children, value, index, ...other} = props;
    return (
        <Grid
            item={ true }
            xs={ 12 }
            style={ {maxWidth: 800} }
            role="tabpanel"
            hidden={ value !== index }
            id={ `simple-tabpanel-${ index }` }
            aria-labelledby={ `simple-tab-${ index }` }
            { ...other }
        >
            { value === index && children }
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

    return (apps === undefined || apps.length === 0 ? null : (<List sx={ {width: '100%'} } subheader={ <ListSubheader
            sx={ {backgroundColor: '#fafafa'} }>{ title }</ListSubheader> }>
            { apps.map(app => (
                <ListItem key={`app-liste-item-${app.ID}`} onClick={() => { onSelect(app.ID) }}>
                    <ListItemAvatar>
                        <Avatar>
                            { icon }
                        </Avatar>
                    </ListItemAvatar>
                    <ListItemText primary={ app.Name } secondary={ app.Identifier }/>
                </ListItem>
            )) }
        </List>)
    )
}

const SettingsPage: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const {appState, dispatch} = props;

    let params = useParams();

    const navigate = useNavigate();

    function a11yProps(index: number): Map<string, string> {
        return new Map([
            ['id', `simple-tab-${ index }`],
            ['aria-controls', `simple-tabpanel-${ index }`],
        ]);
    }

    const [value, setValue] = useState(0);
    const handleChange = (event: React.ChangeEvent<{}>, newValue: number): void => {
        setValue(newValue);
    };

    const [accessTokens, setAccessTokens] = useState<IAccessTokenData[]>([]);

    useEffect(() => {
        getAccessTokens(params.project_id as string).then(response => {
            setAccessTokens(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

    const handleDeleteAccessToken = (accessTokenID: number): void => {
        deleteAccessToken(params.project_id as string, accessTokenID).then(value => {
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
        createAccessToken(params.project_id as string, newToken).then(response => {
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

    const [selectedAppID, setSelectedAppID] = useState<number | null>(appState.project?.Apps === undefined ? null : appState.project?.Apps.length === 0 ? null : appState.project?.Apps[0].ID);
    const selectedApp = appState.project?.Apps.find(a => a.ID === selectedAppID);

    const iosApps = appState.project?.Apps.filter(a => a.Platform === PlatformType.iOS);
    const androidApps = appState.project?.Apps.filter(a => a.Platform === PlatformType.Android);
    const editorApps = appState.project?.Apps.filter(a => a.Platform === PlatformType.Editor);
    const webApps = appState.project?.Apps.filter(a => a.Platform === PlatformType.Web);

    const selectApp = (id:number) => {
        setSelectedAppID(id);
    };

    return (
        <Grid container={ true } spacing={ 2 }>
            <Grid item={ true } xs={ 12 }>
                <Typography variant={ "h1" }>Project Settings</Typography>
            </Grid>
            <Grid item={ true } xs={ 12 }>
                <Tabs
                    value={ value }
                    onChange={ handleChange }
                    indicatorColor="primary"
                    textColor="inherit"
                    aria-label="secondary tabs"
                >
                    <Tab label="General" { ...a11yProps(0) } />
                    <Tab label="Access Tokens" { ...a11yProps(1) } />
                    <Tab label="Users and Permissions" { ...a11yProps(2) } />
                    <Tab label="Notifications" { ...a11yProps(3) } />
                </Tabs>
                <Divider/>
            </Grid>
            <Grid item={ true } container={ true } xs={ 12 } alignItems={ "center" } justifyContent={ "center" }>
                { /* General */ }
                <TabPanel index={ value } value={ 0 }>
                    <TitleCard title={ "Project" }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
                            <Box sx={ {width: '100%', padding: 2} }>
                                <Grid container={ true } spacing={ 4 }>
                                    <Grid item={ true } xs={ 3 }>
                                        <Typography variant={ "body1" } color={ "dimgray" }>Project
                                            name</Typography>
                                    </Grid>
                                    <Grid item={ true } xs={ 9 }>
                                        { appState.project?.Name }
                                    </Grid>
                                    <Grid item={ true } xs={ 3 }>
                                        <Typography variant={ "body1" } color={ "dimgray" }>Project
                                            ID</Typography>
                                    </Grid>
                                    <Grid item={ true } xs={ 9 }>
                                        { appState.project?.Identifier }
                                    </Grid>
                                </Grid>
                            </Box>
                        </Paper>
                    </TitleCard>
                    <TitleCard title={ "Apps" }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
                            <Grid container={ true }>
                                <Grid item={ true } xs={ 12 } container={ true } justifyContent={ "flex-end" } sx={ {
                                    padding: 1,
                                    backgroundColor: '#fafafa',
                                    borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                                } }>
                                    <Button variant={ "contained" }>Add App</Button>
                                </Grid>
                                <Grid item={ true } container={ true } xs={ 4 }
                                      sx={ {backgroundColor: '#fafafa', borderRight: '1px solid rgba(0, 0, 0, 0.12)'} }>
                                    <AppNavigation title={ "Android apps" } apps={ androidApps }
                                                   onSelect={ selectApp } icon={ <Android/> }/>
                                    <AppNavigation title={ "Apple apps" } apps={ iosApps } onSelect={ selectApp } icon={ <Apple/> }/>
                                    <AppNavigation title={ "Unity Editor" } apps={ editorApps }
                                                   onSelect={ selectApp } icon={ <Web/> }/>
                                    <AppNavigation title={ "Web apps" } apps={ webApps } onSelect={ selectApp } icon={ <Web/> }/>
                                </Grid>
                                <Grid item={ true } container={ true } xs={ 8 } sx={ {padding: 2} }>
                                    {
                                        selectedApp === null ? (<Typography variant={ "body1" } color={ "dimgray" }>No
                                            Apps</Typography>) : (<>
                                            <Grid item={ true } xs={ 12 }>
                                                <Typography variant={ "caption" }>App ID</Typography>
                                            </Grid>
                                            <Grid item={ true } xs={ 12 }>
                                                { selectedApp?.ID }
                                            </Grid>
                                            <Grid item={ true } xs={ 12 }>
                                                <Typography variant={ "caption" }>App Name</Typography>
                                            </Grid>
                                            <Grid item={ true } xs={ 12 }>
                                                { selectedApp?.Name }
                                            </Grid>
                                            <Grid item={ true } xs={ 12 }>
                                                <Typography variant={ "caption" }>Bundle Identifier</Typography>
                                            </Grid>
                                            <Grid item={ true } xs={ 12 }>
                                                { selectedApp?.Identifier }
                                            </Grid>
                                            <Grid item={ true } xs={ 12 }>
                                                <Typography variant={ "caption" }>Default Parameter</Typography>
                                            </Grid>
                                            <Grid item={ true } xs={ 12 }>
                                                { selectedApp?.DefaultParameter }
                                            </Grid>
                                        </>)
                                    }
                                </Grid>
                            </Grid>
                        </Paper>
                    </TitleCard>
                </TabPanel>
                { /* Access Tokens */ }
                <TabPanel index={ value } value={ 1 }>
                    <TitleCard title={ "Access Tokens" }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
                            <Box sx={ {width: '100%'} }>
                                <Box sx={ {p: 3} }>
                                    <Table size="small" aria-label="a dense table">
                                        <TableHead>
                                            <TableRow>
                                                <TableCell>Name</TableCell>
                                                <TableCell>Token</TableCell>
                                                <TableCell align="right">Expires</TableCell>
                                                <TableCell></TableCell>
                                            </TableRow>
                                        </TableHead>
                                        <TableBody>
                                            { accessTokens.map((accessToken) => <TableRow key={ accessToken.ID }>
                                                <TableCell>{ accessToken.Name }</TableCell>
                                                <TableCell>
                                                    { accessToken.Token }
                                                    <CopyToClipboard>
                                                        { ({copy}) => (
                                                            <IconButton
                                                                color={ 'primary' }
                                                                size={ 'small' }
                                                                onClick={ () => copy(accessToken.Token) }
                                                            >
                                                                <ContentCopy/>
                                                            </IconButton>
                                                        ) }
                                                    </CopyToClipboard>
                                                </TableCell>
                                                <TableCell align="right">{ accessToken.ExpiresAt !== null ? (<Moment
                                                    format="YYYY/MM/DD HH:mm:ss">{ accessToken.ExpiresAt }</Moment>) : ('Unlimited') }</TableCell>
                                                <TableCell>
                                                    <IconButton color="secondary" size="small" onClick={ () => {
                                                        handleDeleteAccessToken(accessToken.ID as number);
                                                    } }>
                                                        <DeleteForeverIcon/>
                                                    </IconButton>
                                                </TableCell>
                                            </TableRow>) }

                                            <TableRow>
                                                <TableCell>
                                                    <TextField id="new_name" label="Name" variant="outlined"
                                                               value={ newToken.Name } size="small"
                                                               onChange={ event => setNewToken(prevState => ({
                                                                   ...prevState,
                                                                   Name: event.target.value
                                                               })) }/>
                                                </TableCell>
                                                <TableCell></TableCell>
                                                <TableCell align="right">
                                                    <LocalizationProvider dateAdapter={ AdapterDayjs }>
                                                        <MobileDatePicker
                                                            label="Expires At"
                                                            inputFormat="MM/DD/YYYY"
                                                            value={ newToken.ExpiresAt }
                                                            onChange={ (newValue: Dayjs | null, keyvalue: string | undefined) => {
                                                                setNewToken(prevState => ({
                                                                    ...prevState,
                                                                    ExpiresAt: newValue
                                                                }))
                                                            } }
                                                            renderInput={ (params) => <TextField { ...params }
                                                                                                 size="small"/> }
                                                        />
                                                    </LocalizationProvider>
                                                </TableCell>
                                                <TableCell>
                                                    <Button variant="contained" color="primary" size="small"
                                                            onClick={ createNewAccessToken }>
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
                { /* User and Permissions */ }
                <TabPanel index={ value } value={ 2 }>
                    <TitleCard title={ "Users and Permission" }>
                        not implemented
                    </TitleCard>
                </TabPanel>
                { /* Notifications */ }
                <TabPanel index={ value } value={ 3 }>
                    <TitleCard title={ "Notifications" }>
                        not implemented
                    </TitleCard>
                </TabPanel>
            </Grid>
        </Grid>
    );
};

export default SettingsPage;
