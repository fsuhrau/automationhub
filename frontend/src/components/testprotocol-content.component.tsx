import React, { ReactElement, useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import Grid from '@mui/material/Grid';
import ITestRunData from '../types/test.run';
import { TestResultState } from '../types/test.result.state.enum';
import ITestProtocolData, { duration } from '../types/test.protocol';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import { AvTimer, DateRange, PhoneAndroid, Speed } from '@mui/icons-material';
import { Box, Button, Card, CardMedia, Popover, Tab, Tabs } from '@mui/material';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import Moment from 'react-moment';
import moment from 'moment';
import IProtocolEntryData from '../types/protocol.entry';
import TestStatusIconComponent from '../components/test-status-icon.component';
import TestStatusTextComponent from '../components/test-status-text.component';
import { useSSE } from 'react-hooks-sse';
import ProtocolLogComponent from './protocol.log.component';
import ProtocolScreensComponent from './protocol.screens.component';
import IProtocolPerformanceEntryData from '../types/protocol.performance.entry';
import { Chart, LineSeries, Tooltip, ValueAxis } from '@devexpress/dx-react-chart-material-ui';
import { makeStyles } from '@mui/styles';

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
}

function TabPanel(props: TabPanelProps): ReactElement {
    const { children, value, index, ...other } = props;

    return (
        <div
            role="tabpanel"
            hidden={ value !== index }
            id={ `simple-tabpanel-${ index }` }
            aria-labelledby={ `simple-tab-${ index }` }
            { ...other }
        >
            { value === index && (
                <Box sx={ { p: 3 } }>
                    <Typography>{ children }</Typography>
                </Box>
            ) }
        </div>
    );
}

function a11yProps(index: number): Map<string, string> {
    return new Map([
        ['id', `simple-tab-${ index }`],
        ['aria-controls', `simple-tabpanel-${ index }`],
    ]);
}

interface NewTestProtocolLogPayload {
    TestProtocolID: number,
    Entry: IProtocolEntryData,
}

interface TestProtocolContentProps {
    run: ITestRunData
    protocol: ITestProtocolData
}

const TestProtocolContent: React.FC<TestProtocolContentProps> = (props) => {
    const { run, protocol } = props;

    const [anchorScreenEl, setAnchorScreenEl] = useState<HTMLButtonElement | null>(null);
    const showScreenPopup = (event: React.MouseEvent<HTMLButtonElement>): void => {
        setAnchorScreenEl(event.currentTarget);
    };
    const hideScreenPopup = (): void => {
        setAnchorScreenEl(null);
    };

    const lastScreenOpen = Boolean(anchorScreenEl);
    const lastScreenID = lastScreenOpen ? 'simple-popover' : undefined;

    const [lastScreen, setLastScreen] = useState<IProtocolEntryData>();
    const [lastError, setLastError] = useState<IProtocolEntryData>();
    const [lastStep, setLastStep] = useState<IProtocolEntryData>();

    const [entries, setEntries] = useState<IProtocolEntryData[]>([]);
    const [screenEntries, setScreenEntries] = useState<IProtocolEntryData[]>([]);
    const [performanceEntries, setPerformanceEntries] = useState<IProtocolPerformanceEntryData[]>([]);

    const [steps, setSteps] = useState<number>(0);
    const [startupTime, setStartupTime] = useState<number>(0);
    const [fps, setFps] = useState<number>(0);
    const [memory, setMemory] = useState<number>(0);
    const [cpu, setCpu] = useState<number>(0);

    const protocolEntry = useSSE<NewTestProtocolLogPayload | null>(`test_protocol_${ protocol.ID }_log`, null);
    useEffect(() => {
        if (protocolEntry === null)
            return;

        setEntries(prevState => {
            const newState = [...prevState];
            newState.push(protocolEntry.Entry);
            return newState;
        });
    }, [protocolEntry]);

    const updateStatusEntries = (): void => {
        const length = entries.length;
        let numSteps = 0;
        for (let i = 0; i < length; i++) {
            if (lastScreen === undefined && entries[ i ].Source === 'screen') {
                setLastScreen(entries[ i ]);
            }
            if (entries[ i ].Source === 'step') {
                numSteps++;
                if (lastStep === undefined) {
                    setLastStep(entries[ i ]);
                }
            }
            if (lastError === undefined && entries[ i ].Level === 'error') {
                setLastError(entries[ i ]);
            }
        }
        setSteps(numSteps);
    };

    const updatePerformanceStats = (): void => {
        const length = performanceEntries.length;
        if (length === 0) {
            return;
        }
        let sumFps = 0;
        let sumMem = 0;
        let sumCpu = 0;
        performanceEntries.forEach(value1 => {
            sumCpu += value1.CPU;
            sumFps += value1.FPS;
            sumMem += value1.MEM;
        });
        setFps(sumFps / length);
        setMemory(sumMem / length);
        setCpu(sumFps / length);
    };

    const [value, setValue] = useState(0);
    const handleChange = (event: React.ChangeEvent<{}>, newValue: number): void => {
        setValue(newValue);
    };

    useEffect(() => {
        setEntries(protocol.Entries);
        setPerformanceEntries(protocol.Performance);
    }, [protocol]);

    useEffect(() => {
        updateStatusEntries();
        setScreenEntries(entries.filter(value1 => value1.Source === 'screen'));
    }, [entries]);

    useEffect(() => {
        updatePerformanceStats();
    }, [performanceEntries]);

    return (
        <Paper sx={{ maxWidth: 1200, margin: 'auto', overflow: 'hidden' }}>
            <AppBar
                position="static"
                color="default"
                elevation={0}
                sx={{ borderBottom: '1px solid rgba(0, 0, 0, 0.12)' }}
            >
                <Toolbar>
                    <Grid container={ true } spacing={ 2 } alignItems="center">
                        <Grid item={ true }>
                            <TestStatusIconComponent status={ protocol.TestResult }/>
                        </Grid>
                        <Grid item={ true }>
                            <TestStatusTextComponent status={ protocol.TestResult }/>
                        </Grid>

                        <Grid item={ true }>
                            <DateRange sx={{ display: 'block' }} color="inherit"/>
                        </Grid>
                        <Grid item={ true }>
                            <Moment format="YYYY/MM/DD HH:mm:ss">{ protocol.StartedAt }</Moment>
                        </Grid>

                        <Grid item={ true }>
                            <AvTimer sx={{ display: 'block' }} color="inherit"/>
                        </Grid>
                        <Grid item={ true }>
                            { duration(protocol.StartedAt, protocol.EndedAt) }
                        </Grid>

                        <Grid item={ true }>
                            <PhoneAndroid sx={{ display: 'block' }} color="inherit"/>
                        </Grid>
                        <Grid item={ true }>
                            { protocol.Device?.Name }
                        </Grid>

                        <Grid item={ true }>
                            <Speed sx={{ display: 'block' }} color="inherit"/>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                            { protocol.TestName }
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
                            indicatorColor="primary"
                            textColor="inherit"
                        >
                            <Tab label="Status" { ...a11yProps(0) } />
                            <Tab label="Executor" { ...a11yProps(1) } />
                            <Tab label="Logs" { ...a11yProps(2) } />
                            <Tab label="Screenshots" { ...a11yProps(3) } />
                            <Tab label="Video/Replay" { ...a11yProps(4) } />
                            <Tab label="Performance" { ...a11yProps(5) } />
                        </Tabs>
                    </Box>
                </AppBar>
                <TabPanel value={ value } index={ 0 }>
                    <Grid container={ true } direction="column" spacing={ 2 }>
                        { protocol?.TestResult == TestResultState.TestResultFailed &&
                        <Box width='100%' color="white" bgcolor="palevioletred" padding={ 2 } margin={-1}>
                            <Grid container={ true } spacing={ 6 } justifyContent="center" alignItems="center">
                                <Grid item={ true }>
                                    { lastStep && <Typography><Moment
                                        format="YYYY/MM/DD HH:mm:ss">{ lastStep.CreatedAt }</Moment>: { lastStep.Message }
                                    </Typography> }
                                    { lastError && <Typography><Moment
                                        format="YYYY/MM/DD HH:mm:ss">{ lastError.CreatedAt }</Moment>: { lastError.Message }
                                    </Typography> }
                                </Grid>
                                <Grid item={ true }>
                                    { lastScreen && <div>
                                        <Button aria-describedby={ 'last_screen_' + lastScreen.ID }
                                            variant="contained" onClick={ showScreenPopup }>
                                            Show
                                        </Button>
                                        <Popover
                                            id={ lastScreenID }
                                            open={ lastScreenOpen }
                                            anchorEl={ anchorScreenEl }
                                            onClose={ hideScreenPopup }
                                            anchorOrigin={ {
                                                vertical: 'bottom',
                                                horizontal: 'left',
                                            } }
                                        >
                                            <Card>
                                                <CardMedia
                                                    component="img"
                                                    height="400"
                                                    image={ `/api/data/${ lastScreen.Data }` }
                                                    alt="green iguana"
                                                />
                                            </Card>
                                        </Popover>
                                    </div> }
                                </Grid>
                            </Grid>
                        </Box>
                        }
                        <Grid item={ true }>
                            <Grid container={ true } spacing={ 0 }>
                            </Grid>
                        </Grid>
                        <Grid item={ true }>
                            <Grid container={ true } spacing={ 6 }>
                                <Grid item={ true }>
                                    <Grid item={ true } xs={ true } container={ true } direction="column" spacing={ 4 }>
                                        <Grid item={ true } xs={ true }>
                                            <Typography gutterBottom={ true } variant="subtitle1">
                                                Time
                                            </Typography>
                                            <Typography gutterBottom={ true } variant="body2" color="textSecondary">
                                                Execution
                                            </Typography>
                                            <Typography variant="h5" style={ { whiteSpace: 'nowrap' } }>
                                                { duration(protocol.StartedAt, protocol.EndedAt) }
                                            </Typography>
                                        </Grid>
                                    </Grid>
                                </Grid>
                                <Grid item={ true } xs={ 12 } sm={ true } container={ true }>
                                    <Grid item={ true } xs={ true } container={ true } direction="column" spacing={ 2 }>
                                        <Grid item={ true } xs={ true }>
                                            <Typography gutterBottom={ true } variant="subtitle1">
                                                Details
                                            </Typography>
                                            <Grid container={ true } spacing={ 2 }>
                                                <Grid item={ true } xs={ 2 }>
                                                    <Typography gutterBottom={ true } variant="body2"
                                                        color="textSecondary">
                                                        Steps
                                                    </Typography>
                                                    <Typography variant="h5">
                                                        { steps }
                                                    </Typography>
                                                </Grid>
                                                <Grid item={ true } xs={ 2 }>
                                                    <Typography gutterBottom={ true } variant="body2"
                                                        color="textSecondary">
                                                        StartupTime
                                                    </Typography>
                                                    <Typography variant="h5">
                                                        N/A
                                                    </Typography>
                                                </Grid>
                                                <Grid item={ true } xs={ 2 }>
                                                    <Typography gutterBottom={ true } variant="body2"
                                                        color="textSecondary">
                                                        FPS
                                                    </Typography>
                                                    <Typography variant="h5">
                                                        { fps.toFixed(2) }
                                                    </Typography>
                                                </Grid>
                                                <Grid item={ true } xs={ 2 }>
                                                    <Typography gutterBottom={ true } variant="body2"
                                                        color="textSecondary">
                                                        Memory
                                                    </Typography>
                                                    <Typography variant="h5" component="h5">
                                                        { (memory * 1024 * 1024).toFixed(2) }MB
                                                    </Typography>
                                                </Grid>
                                            </Grid>
                                        </Grid>
                                    </Grid>
                                </Grid>
                            </Grid>
                        </Grid>
                    </Grid>
                </TabPanel>
                <TabPanel value={ value } index={ 1 }>
                    <Table size="small" aria-label="a dense table">
                        <TableHead>
                            <TableRow>
                                <TableCell>Date</TableCell>
                                <TableCell>Level</TableCell>
                                <TableCell>Log</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            { run.Log.map((entry) => <TableRow key={ entry.ID }>
                                <TableCell component="th" scope="row" style={ { whiteSpace: 'nowrap' } }>
                                    <Moment format="YYYY/MM/DD HH:mm:ss">{ entry.CreatedAt }</Moment>
                                </TableCell>
                                <TableCell>{ entry.Level }</TableCell>
                                <TableCell>{ entry.Log }</TableCell>
                            </TableRow>) }
                        </TableBody>
                    </Table>
                </TabPanel>
                <TabPanel value={ value } index={ 2 }>
                    <ProtocolLogComponent entries={ entries } />
                </TabPanel>
                <TabPanel value={ value } index={ 3 }>
                    <ProtocolScreensComponent entries={ screenEntries }/>
                </TabPanel>
                <TabPanel value={ value } index={ 4 }>
                    Video (not implemented)
                </TabPanel>
                <TabPanel value={ value } index={ 5 }>
                    <Grid item={ true } xs={ true } container={ true } direction="column" spacing={ 4 }>
                        <Grid item={ true } xs={ true }>
                            <Typography gutterBottom={ true } variant="subtitle1">
                                FPS
                            </Typography>
                            <Chart
                                data={ performanceEntries }
                            >
                                <ValueAxis/>
                                <LineSeries valueField="FPS" argumentField="Runtime"/>
                                <Tooltip/>
                            </Chart>
                        </Grid>
                        <Grid item={ true } xs={ true }>
                            <Typography gutterBottom={ true } variant="subtitle1">
                                Memory
                            </Typography>
                            <Chart
                                data={ performanceEntries }
                            >
                                <ValueAxis/>
                                <LineSeries valueField="MEM" argumentField="Runtime"/>
                                <Tooltip/>
                            </Chart>
                        </Grid>
                        { cpu > 0 && (<Grid item={ true } xs={ true }>
                            <Typography gutterBottom={ true } variant="subtitle1">
                                CPU
                            </Typography>
                            <Chart
                                data={ performanceEntries }
                                >
                                <ValueAxis/>
                                <LineSeries valueField="CPU" argumentField="Runtime"/>
                                <Tooltip/>
                            </Chart>
                            </Grid>
                        ) }
                    </Grid>
                </TabPanel>
            </Box>
        </Paper>
    );
};

export default TestProtocolContent;
