import React from 'react';
import Paper from '@mui/material/Paper';
import { Alert, Box, Button, Divider, Grid, Typography } from '@mui/material';
import { getTestExecutionName } from '../../types/test.execution.type.enum';
import { getTestTypeName, TestType } from '../../types/test.type.enum';
import { useHistory } from 'react-router-dom';
import ITestData from '../../types/test';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import { getPlatformName } from "../../types/platform.type.enum";

function getUnityTestsConfig(): Array<Object> {
    return [{id: 0, name: 'Run all Tests'}, {id: 1, name: 'Run only Selected Tests'}];
}

function getDeviceOption(): Array<Object> {
    return [{id: 0, name: 'All Devices'}, {id: 1, name: 'Selected Devices Only'}];
}

interface TestContentProps {
    test: ITestData
}

const ShowTestPage: React.FC<TestContentProps> = (props) => {

    const {test} = props;
    const history = useHistory();

    const unityTestConfig = getUnityTestsConfig();
    const deviceConfig = getDeviceOption();

    const getUnityTestConfigName = (b: boolean): string => {
        const id = b ? 0 : 1;
        const item = unityTestConfig.find(i => i.id == id);
        return item.name;
    };

    const getDeviceConfigName = (b: boolean): string => {
        const id = b ? 0 : 1;
        const item = deviceConfig.find(i => i.id == id);
        return item.name;
    };

    return (
        <Paper sx={ {maxWidth: 1200, margin: 'auto', overflow: 'hidden'} }>
            <AppBar
                position="static"
                color="default"
                elevation={ 0 }
                sx={ {borderBottom: '1px solid rgba(0, 0, 0, 0.12)'} }
            >
                <Toolbar>
                    <Grid container={ true } spacing={ 2 } alignItems="center">
                        <Grid item={ true }>
                            <Typography variant={ 'h6' }>
                                Test: { test.Name }
                            </Typography>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                        </Grid>
                        <Grid item={ true }>
                            <Button variant="contained" color="primary" size="small"
                                    href={ `${ test.ID }/edit` }>Edit</Button>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <Box sx={ {p: 2, m: 2} }>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h6' }>Test Data</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            <Grid item={ true } xs={ 2 }>
                                Type:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { getTestTypeName(test.TestConfig.Type) }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Execution:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { getTestExecutionName(test.TestConfig.ExecutionType) }
                                <br/>
                                <Alert severity="info">
                                    Concurrent = runs each test on a different free
                                    device to get faster results<br/>
                                    Simultaneously = runs every test on every device
                                    to get a better accuracy</Alert>
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Platform:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { getPlatformName(test.TestConfig.Platform) }
                            </Grid>
                            <Grid item={ true } xs={ 2 }>
                                Devices:
                            </Grid>
                            <Grid item={ true } xs={ 10 }>
                                { getDeviceConfigName(test.TestConfig.AllDevices) }
                                { test.TestConfig.AllDevices === true && (<div>all</div>) }
                                { test.TestConfig.AllDevices === false && (<div>
                                    Devices:<br/>
                                    { test.TestConfig.Devices.map((a) =>
                                        <div>- { a.Device?.DeviceIdentifier }({ a.Device?.Name })<br/></div>,
                                    ) }
                                </div>) }
                            </Grid>
                        </Grid>

                        <Typography variant={ 'h6' }>Config</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true }>
                            { test.TestConfig.Type === TestType.Unity && (
                                <>
                                    <Grid item={ true } xs={ 2 }>
                                        Run:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { getUnityTestConfigName(test.TestConfig.Unity?.RunAllTests) }
                                    </Grid>
                                    <Grid item={ true } xs={ 2 }>
                                        Categories
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { test.TestConfig.Unity?.Categories }
                                    </Grid>
                                    { test.TestConfig.Unity?.RunAllTests === false && (
                                        <>
                                            <Grid item={ true } xs={ 2 }>
                                                Functions:
                                            </Grid>
                                            <Grid item={ true } xs={ 10 }>
                                                Functions:<br/>
                                                { test.TestConfig.Unity.UnityTestFunctions.map((a) =>
                                                    <div>- { a.Class }/{ a.Method }<br/></div>,
                                                ) }
                                            </Grid>
                                        </>) }
                                </>) }
                        </Grid>
                    </Grid>
                </Grid>
            </Box>
        </Paper>
    );
};

export default ShowTestPage;
