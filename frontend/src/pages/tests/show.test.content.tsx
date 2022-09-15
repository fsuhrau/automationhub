import React from 'react';
import Paper from '@mui/material/Paper';
import { Alert, Box, Button, Divider, Grid, Typography } from '@mui/material';
import { getTestExecutionName } from '../../types/test.execution.type.enum';
import { getTestTypeName, TestType } from '../../types/test.type.enum';
import { useNavigate } from 'react-router-dom';
import ITestData from '../../types/test';
import { TitleCard } from "../../components/title.card.component";

type KeyValue = { id: number, name: string };

function getUnityTestsConfig(): Array<KeyValue> {
    return [{id: 0, name: 'Run all Tests'}, {id: 1, name: 'Run only Selected Tests'}];
}

function getDeviceOption(): Array<KeyValue> {
    return [{id: 0, name: 'All Devices'}, {id: 1, name: 'Selected Devices Only'}];
}

interface TestContentProps {
    test: ITestData
}

const ShowTestPage: React.FC<TestContentProps> = (props) => {

    const {test} = props;

    const history = useNavigate();

    const unityTestConfig = getUnityTestsConfig();
    const deviceConfig = getDeviceOption();

    const getUnityTestConfigName = (b: boolean): string => {
        const id = b ? 0 : 1;
        const item = unityTestConfig.find(i => i.id === id);
        return item === undefined ? "" : item.name;
    };

    const getDeviceConfigName = (b: boolean): string => {
        const id = b ? 0 : 1;
        const item = deviceConfig.find(i => i.id === id);
        return item === undefined ? "" : item.name;
    };

    return (
        <Grid container={ true } spacing={ 2 }>
            <Grid item={ true } xs={ 12 }>
                <Typography variant={ "h1" }>Test -AppID-</Typography>
            </Grid>
            <Grid item={ true } xs={ 12 }>
                <Divider/>
            </Grid>
            <Grid item={ true } container={ true } xs={ 12 } alignItems={ "center" } justifyContent={ "center" }>
                <Grid
                    item={ true }
                    xs={ 12 }
                    style={ {maxWidth: 800} }
                >
                    <TitleCard title={ 'Test' }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
                            <Grid container={ true }>
                                <Grid item={ true } xs={ 6 } container={ true } sx={ {
                                    padding: 2,
                                    backgroundColor: '#fafafa',
                                    borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                                } }>
                                    { test.Name }
                                </Grid>
                                <Grid item={ true } xs={ 6 } container={ true } justifyContent={ "flex-end" } sx={ {
                                    padding: 1,
                                    backgroundColor: '#fafafa',
                                    borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                                } }>
                                    <Button variant="contained" color="primary" size="small"
                                            href={ `${ test.ID }/edit` }>Edit</Button>
                                </Grid>
                                <Grid item={ true } xs={ 12 }>
                                    <Box sx={ {p: 2, m: 2} }>
                                        <Grid container={ true }>
                                            <Grid item={ true } container={ true } xs={ 12 } spacing={ 2 }>
                                                <Grid item={ true } xs={ 12 }>
                                                    <Typography variant={ 'h6' }>Test Data</Typography>
                                                </Grid>
                                                <Grid item={ true } xs={ 12 }>
                                                    <Divider/>
                                                </Grid>
                                                <Grid item={ true } xs={ 2 }>
                                                    Type:
                                                </Grid>
                                                <Grid item={ true } xs={ 10 }>
                                                    { getTestTypeName(test.TestConfig.Type) }
                                                </Grid>
                                                <Grid item={ true } xs={ 2 }>
                                                    Execution:
                                                </Grid>
                                                <Grid item={ true } container={ true } xs={ 10 } spacing={ 2 }>
                                                    <Grid item={ true } xs={ 12 }>
                                                        { getTestExecutionName(test.TestConfig.ExecutionType) }
                                                    </Grid>
                                                    <Grid item={ true } xs={ 12 }>
                                                        <Alert severity="info">
                                                            Concurrent = runs each test on a different free
                                                            device to get faster results<br/>
                                                            Simultaneously = runs every test on every device
                                                            to get a better accuracy</Alert>
                                                    </Grid>
                                                    <Grid item={ true } xs={ 2 }>
                                                        Devices:
                                                    </Grid>
                                                    <Grid item={ true } xs={ 10 }>
                                                        { getDeviceConfigName(test.TestConfig.AllDevices) }
                                                        { test.TestConfig.AllDevices && (<div>all</div>) }
                                                        { !test.TestConfig.AllDevices && (<div>
                                                        </div>) }
                                                    </Grid>
                                                    { !test.TestConfig.AllDevices && (<>
                                                        <Grid item={ true } xs={ 2 }>
                                                            Devices:
                                                        </Grid>
                                                        <Grid item={ true } xs={ 10 }>
                                                            { test.TestConfig.Devices.map((a, index) =>
                                                                <div
                                                                    key={ `device_${ a.ID }_${ index }` }>- { a.Device?.DeviceIdentifier }({ a.Device?.Name })<br/>
                                                                </div>,
                                                            ) }
                                                        </Grid>
                                                    </>) }
                                                </Grid>
                                                <Grid item={ true } container={ true } xs={ 12 } spacing={ 2 }>
                                                    <Grid item={ true } xs={ 12 }>
                                                        <Typography variant={ 'h6' }>Config</Typography>
                                                    </Grid>
                                                    <Grid item={ true } xs={ 12 }>
                                                        <Divider/>
                                                    </Grid>
                                                    { test.TestConfig.Type === TestType.Unity && (
                                                        <>
                                                            <Grid item={ true } xs={ 2 }>
                                                                Run:
                                                            </Grid>
                                                            <Grid item={ true } xs={ 10 }>
                                                                { getUnityTestConfigName(test.TestConfig.Unity?.RunAllTests) }
                                                            </Grid>
                                                            {
                                                                test.TestConfig.Unity?.Categories !== "" && <>
                                                                    <Grid item={ true } xs={ 2 }>
                                                                        Categories
                                                                    </Grid>
                                                                    <Grid item={ true } xs={ 10 }>
                                                                        { test.TestConfig.Unity?.Categories }
                                                                    </Grid>
                                                                </>
                                                            }
                                                            { test.TestConfig.Unity?.RunAllTests === false && (
                                                                <>
                                                                    <Grid item={ true } xs={ 2 }>
                                                                        Functions:
                                                                    </Grid>
                                                                    <Grid item={ true } xs={ 10 }>
                                                                        Functions:<br/>
                                                                        { test.TestConfig.Unity.UnityTestFunctions.map((a, index) =>
                                                                            <div
                                                                                key={ `test_function_${ a.ID }_${ index }` }>- { a.Class }/{ a.Method }<br/>
                                                                            </div>,
                                                                        ) }
                                                                    </Grid>
                                                                </>) }
                                                        </>) }
                                                </Grid>
                                            </Grid>
                                        </Grid>
                                    </Box>
                                </Grid>
                            </Grid>
                        </Paper>
                    </TitleCard>
                </Grid>
            </Grid>
        </Grid>
    );
};

export default ShowTestPage;
