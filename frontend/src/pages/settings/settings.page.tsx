import React, { ChangeEvent, ReactElement, useEffect, useState } from 'react';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Paper from '@mui/material/Paper';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import { useHistory } from 'react-router-dom';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import Moment from 'react-moment';
import { Box, Tab, Tabs, TextField, Typography } from '@mui/material';
import {
    createAccessToken,
    deleteAccessToken,
    getAccessTokens,
    NewAccessTokenRequest,
} from '../../services/settings.service';
import IAccessTokenData from '../../types/access.token';
import CopyToClipboard from '../../components/copy.clipboard.component';
import { ContentCopy } from '@mui/icons-material';
import IconButton from '@mui/material/IconButton';
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { MobileDatePicker, LocalizationProvider, MuiPickersAdapterContext } from '@mui/x-date-pickers';
import dayjs, { Dayjs } from 'dayjs';

const SettingsPage: React.FC = () => {
    interface TabPanelProps {
        children?: React.ReactNode;
        index: number;
        value: number;
    }

    const history = useHistory();

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
        getAccessTokens().then(response => {
            setAccessTokens(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

    const handleDeleteAccessToken = (accessTokenID: number): void => {
        deleteAccessToken(accessTokenID).then(value => {
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
        createAccessToken(newToken).then(response => {
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

    return (
        <>
            <Paper sx={ { maxWidth: 1200, margin: 'auto', overflow: 'hidden' } }>
                <AppBar
                    position="static"
                    color="default"
                    elevation={ 0 }
                    sx={ { borderBottom: '1px solid rgba(0, 0, 0, 0.12)' } }
                >
                    <Toolbar>
                        <Grid container={ true } spacing={ 2 } alignItems="center">
                            <Grid item={ true }>
                                <Typography variant={ 'h6' }>
                                    Settings
                                </Typography>
                            </Grid>
                            <Grid item={ true } xs={ true }>
                            </Grid>
                            <Grid item={ true }>
                            </Grid>
                        </Grid>
                    </Toolbar>
                </AppBar>
                <Box sx={ { width: '100%' } }>
                    <AppBar position="static" color="default" elevation={ 0 }>
                        <Box sx={ { borderBottom: 1, borderColor: 'divider' } }>
                            <Tabs
                                value={ value }
                                onChange={ handleChange }
                                indicatorColor="secondary"
                                textColor="inherit"
                                aria-label="secondary tabs"
                            >
                                <Tab label="Notifications" { ...a11yProps(0) } />
                                <Tab label="AccessTokens" { ...a11yProps(1) } />
                            </Tabs>
                        </Box>
                    </AppBar>
                    <div
                        role="tabpanel"
                        hidden={ value !== 0 }
                        id={ `simple-tabpanel-${ 0 }` }
                        aria-labelledby={ `simple-tab-${ 0 }` }
                    >
                        { value === 0 && (
                            <Box sx={ { p: 3 } }>
                                <Typography>Todo</Typography>
                            </Box>
                        ) }
                    </div>
                    <div
                        role="tabpanel"
                        hidden={ value !== 1 }
                        id={ `simple-tabpanel-${ 1 }` }
                        aria-labelledby={ `simple-tab-${ 1 }` }
                    >
                        { value === 1 && (
                            <Box sx={ { p: 3 } }>
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
                                                    {({ copy }) => (
                                                        <IconButton
                                                            color={'primary'}
                                                            size={ 'small' }
                                                            onClick={() => copy(accessToken.Token)}
                                                        >
                                                            <ContentCopy />
                                                        </IconButton>
                                                    )}
                                                </CopyToClipboard>
                                            </TableCell>
                                            <TableCell align="right">{accessToken.ExpiresAt !== null ? (<Moment format="YYYY/MM/DD HH:mm:ss">{ accessToken.ExpiresAt }</Moment>) : ('Unlimited') }</TableCell>
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
                                                <TextField id="new_name" label="Name" variant="outlined" value={ newToken.Name } size="small" onChange={ event => setNewToken(prevState => ({...prevState, Name: event.target.value}))} />
                                            </TableCell>
                                            <TableCell></TableCell>
                                            <TableCell align="right">
                                                <LocalizationProvider dateAdapter={AdapterDayjs}>
                                                    <MobileDatePicker
                                                        label="Expires At"
                                                        inputFormat="MM/DD/YYYY"
                                                        value={newToken.ExpiresAt}
                                                        onChange={(newValue: Dayjs | null, keyvalue: string | undefined) => {
                                                            setNewToken(prevState => ({...prevState, ExpiresAt: newValue }))
                                                        }}
                                                        renderInput={(params) => <TextField {...params} size="small"  />}
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
                        ) }
                    </div>
                </Box>
            </Paper>
        </>
    );
};

export default SettingsPage;
