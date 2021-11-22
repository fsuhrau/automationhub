import React, { FC } from 'react';
import Paper from '@material-ui/core/Paper';
import { createStyles, Theme, WithStyles, withStyles } from '@material-ui/core/styles';
import { Box, Button, Divider, Grid, Typography } from '@material-ui/core';
import { TestExecutionType } from '../../types/test.execution.type.enum';
import { TestType } from '../../types/test.type.enum';
import { useHistory } from 'react-router-dom';
import ITestData from '../../types/test';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';

const styles = (theme: Theme): ReturnType<typeof createStyles> =>
    createStyles({
        paper: {
            maxWidth: 1200,
            margin: 'auto',
            overflow: 'hidden',
        },
        searchBar: {
            borderBottom: '1px solid rgba(0, 0, 0, 0.12)',
        },
        searchInput: {
            fontSize: theme.typography.fontSize,
        },
        root: {
            width: '100%',
        },
        backButton: {
            marginRight: theme.spacing(1),
        },
        instructions: {
            marginTop: theme.spacing(1),
            marginBottom: theme.spacing(1),
        },
        formControl: {
            margin: theme.spacing(1),
            minWidth: 120,
        },
        selectEmpty: {
            marginTop: theme.spacing(2),
        },
        contentWrapper: {
            margin: '40px 16px',
        },
    });

const StringIsNumber = (value: any): boolean => !isNaN(Number(value));

function ToArray(en: any): Array<Object> {
    return Object.keys(en).filter(StringIsNumber).map(key => ({ id: key, name: en[ key ] }));
}

function getExecutionTypes(): Array<Object> {
    return ToArray(TestExecutionType);
}

function getTestTypes(): Array<Object> {
    return ToArray(TestType);
}

function getUnityTestsConfig(): Array<Object> {
    return [{ id: 0, name: 'Run all Tests' }, { id: 1, name: 'Run only Selected Tests' }];
}

function getDeviceOption(): Array<Object> {
    return [{ id: 0, name: 'All Devices' }, { id: 1, name: 'Selected Devices Only' }];
}

interface TestContentProps extends WithStyles<typeof styles> {
    test: ITestData
}


const ShowTestPage: FC<TestContentProps> = (props) => {
    const { test, classes } = props;
    const history = useHistory();

    const testTypes = getTestTypes();
    const executionTypes = getExecutionTypes();
    const unityTestConfig = getUnityTestsConfig();
    const deviceConfig = getDeviceOption();

    const getTestTypeName = (type: TestType): string => {
        const item = testTypes.find(i => i.id == type);
        return item.name;
    };

    const getTestExecutionName = (type: TestExecutionType): string => {
        const item = executionTypes.find(i => i.id == type);
        return item.name;
    };

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
        <Paper className={ classes.paper }>
            <AppBar className={ classes.searchBar } position="static" color="default" elevation={ 0 }>
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
            <Box sx={ { p: 2, m: 2 } }>
                <Grid container={ true }>
                    <Grid item={ true } xs={ 12 }>
                        <Typography variant={ 'h6' }>Test Configuration</Typography>
                        <Divider/>
                        <br/>
                        <Grid container={ true } >
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
                                <Typography variant={'caption'}>
                                    Concurrent = runs each test on a different free
                                    device to get faster results<br/>
                                    Simultaneously = runs every test on every device
                                    to get a better accuracy
                                </Typography>
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

                        { test.TestConfig.Type === TestType.Unity && (
                            <div>
                                <br/>
                                <Typography variant={ 'h6' }>Unity Config</Typography>
                                <Divider/>
                                <br/>
                                <Grid container={ true }>
                                    <Grid item={ true } xs={ 2 }>
                                        Selected Test Functions:
                                    </Grid>
                                    <Grid item={ true } xs={ 10 }>
                                        { getUnityTestConfigName(test.TestConfig.Unity?.RunAllTests) }
                                        { test.TestConfig.Unity?.RunAllTests === false && (<div>
                                            Functions:<br/>
                                            { test.TestConfig.Unity.UnityTestFunctions.map((a) =>
                                                <div>- { a.Class }/{ a.Method }<br/></div>,
                                            ) }
                                        </div>) }
                                    </Grid>
                                </Grid>
                            </div>) }
                    </Grid>
                </Grid>
            </Box>
        </Paper>
    );
};

export default withStyles(styles)(ShowTestPage);
