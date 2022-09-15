import React, { FunctionComponent, PureComponent, ReactElement, useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import Grid from '@mui/material/Grid';
import ITestRunData from '../types/test.run';
import { TestResultState } from '../types/test.result.state.enum';
import ITestProtocolData, { duration } from '../types/test.protocol';
import Typography from '@mui/material/Typography';
import { AvTimer, DateRange, PhoneAndroid, Speed } from '@mui/icons-material';
import { Button, Card, CardMedia, Divider, Popover, Tab, Tabs } from '@mui/material';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableBody from '@mui/material/TableBody';
import moment from 'moment';
import IProtocolEntryData from '../types/protocol.entry';
import TestStatusIconComponent from '../components/test-status-icon.component';
import { useSSE } from 'react-hooks-sse';
import ProtocolLogComponent from './protocol.log.component';
import ProtocolScreensComponent from './protocol.screens.component';
import IProtocolPerformanceEntryData from '../types/protocol.performance.entry';
import {
    CartesianGrid,
    LabelList,
    Legend,
    Line,
    LineChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis
} from "recharts";
import { TitleCard } from "./title.card.component";

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
}

function TabPanel(props: TabPanelProps): ReactElement {
    const {children, value, index, ...other} = props;

    return (
        <div
            role="tabpanel"
            hidden={ value !== index }
            id={ `simple-tabpanel-${ index }` }
            aria-labelledby={ `simple-tab-${ index }` }
            { ...other }
        >
            { value === index && (
                <>
                    { children }
                </>
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

const CustomLineLabel: FunctionComponent<any> = (props: any) => {
    const { x, y, stroke, value, unit } = props;
    return (
        <text x={x} y={y} dy={-4} fill={stroke} fontSize={10} textAnchor="middle">
            {value.toFixed(2)}{unit !== undefined ? unit : ''}
        </text>
    );
};

const TestProtocolContent: React.FC<TestProtocolContentProps> = (props) => {
    const {run, protocol} = props;

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
    const [lastErrors, setLastErrors] = useState<IProtocolEntryData[]>([]);
    const [lastStep, setLastStep] = useState<IProtocolEntryData>();

    const [entries, setEntries] = useState<IProtocolEntryData[]>([]);
    const [screenEntries, setScreenEntries] = useState<IProtocolEntryData[]>([]);
    const [performanceEntries, setPerformanceEntries] = useState<IProtocolPerformanceEntryData[]>([]);

    const [steps, setSteps] = useState<number>(0);
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
        const errors: IProtocolEntryData[] = [];
        for (let i = length - 1; i > 0; i--) {
            if (lastScreen === undefined && entries[ i ].Source === 'screen') {
                setLastScreen(entries[ i ]);
            }
            if (entries[ i ].Source === 'step') {
                numSteps++;
                if (lastStep === undefined) {
                    setLastStep(entries[ i ]);
                }
            }

            if (entries[ i ].Level === 'error') {
                errors.push(entries[ i ]);
            }
        }

        setLastErrors(errors);
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
    }, [performanceEntries])

    const startupTime = 0;// protocol?.Performance[0].

    const checkpoints = performanceEntries.filter(p => p.Checkpoint !== "schedule").map((value1, index, array) => {
        if (index !== 0) {
            value1.ExecutionTime = value1.Runtime - array[index-1].Runtime
        } else {
            value1.ExecutionTime = value1.Runtime
        }
        return value1
    })

    return (
        <Grid container={ true } spacing={ 2 }>
            <Grid item={ true } xs={ 12 }>
                <Typography variant={ "h1" }><TestStatusIconComponent
                    status={ protocol.TestResult }/>{ ' ' }{ protocol.TestName }</Typography>
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
                    <TitleCard title={ "Test Protocol" }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
                            <Grid container={ true } spacing={ 2 } alignItems="center" sx={ {
                                padding: 1,
                                backgroundColor: protocol?.TestResult == TestResultState.TestResultFailed ? '#ff2b40' : '#05bdae',
                                borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                            } }>
                                <Grid item={ true }>
                                    <DateRange sx={ {display: 'block'} } color="inherit"/>
                                </Grid>
                                <Grid item={ true }>
                                    { moment(protocol.StartedAt).format('YYYY/MM/DD HH:mm:ss') }
                                </Grid>

                                <Grid item={ true }>
                                    <AvTimer sx={ {display: 'block'} } color="inherit"/>
                                </Grid>
                                <Grid item={ true }>
                                    { duration(protocol.StartedAt, protocol.EndedAt) }
                                </Grid>

                                <Grid item={ true }>
                                    <PhoneAndroid sx={ {display: 'block'} } color="inherit"/>
                                </Grid>
                                <Grid item={ true }>
                                    { protocol.Device?.Name }
                                </Grid>
                                <Grid item={ true }>
                                    <Speed sx={ {display: 'block'} } color="inherit"/>
                                </Grid>
                            </Grid>

                            <Grid container={ true }
                                  sx={ {backgroundColor: '#fafafa', borderBottom: '1px solid rgba(0, 0, 0, 0.12)'} }>
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
                            </Grid>
                            <Grid container={ true }>
                                <TabPanel value={ value } index={ 0 }>
                                    <Grid container={ true }>
                                        { protocol?.TestResult == TestResultState.TestResultFailed &&
                                            <Grid item={ true } container={ true } justifyContent="center"
                                                  alignItems="center" sx={ {padding: 1, backgroundColor: '#ff2b40'} }>
                                                <Grid item={ true } xs={ 12 }>
                                                    { lastStep &&
                                                        <Typography>{ moment(lastStep.CreatedAt).format('YYYY/MM/DD HH:mm:ss') }{ ': ' }{ lastStep.Message }</Typography> }
                                                    { lastErrors.map((lastError) => (
                                                        <Typography>{ moment(lastError.CreatedAt).format('YYYY/MM/DD HH:mm:ss') }{ ': ' }{ lastError.Message }</Typography>)) }
                                                </Grid>
                                                <Grid item={ true } xs={ 12 } container={ true }
                                                      justifyContent={ "flex-end" }>
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
                                        }
                                        <Grid item={ true } xs={ 12 } container={ true } sx={ {padding: 1} }>
                                            <Grid item={ true } xs={ 4 } container={ true }>
                                                <Grid item={ true } xs={ 12 }>
                                                    <Typography gutterBottom={ true } variant="subtitle1">
                                                        Time
                                                    </Typography>
                                                </Grid>
                                                <Grid item={ true } xs={ 12 }>
                                                    <Typography gutterBottom={ true } variant="body2" color="textSecondary">
                                                        Execution
                                                    </Typography>
                                                    <Typography variant="h5" style={ {whiteSpace: 'nowrap'} }>
                                                        { duration(protocol.StartedAt, protocol.EndedAt) }
                                                    </Typography>
                                                </Grid>
                                            </Grid>
                                            <Grid item={ true } xs={ 8 } container={ true }>
                                                <Grid item={ true } xs={ 12 }>
                                                    <Typography gutterBottom={ true } variant="subtitle1">
                                                        Details
                                                    </Typography>
                                                </Grid>
                                                <Grid item={ true } container={ true } xs={ 12 } spacing={ 2 }>
                                                    <Grid item={ true } xs={ 3 }>
                                                        <Typography gutterBottom={ true } variant="body2"
                                                                    color="textSecondary">Steps</Typography>
                                                        <Typography variant="h5">{ steps }</Typography>
                                                    </Grid>
                                                    <Grid item={ true } xs={ 3 }>
                                                        <Typography gutterBottom={ true } variant="body2"
                                                                    color="textSecondary">
                                                            StartupTime
                                                        </Typography>
                                                        <Typography variant="h5">
                                                            { startupTime }ms
                                                        </Typography>
                                                    </Grid>
                                                    <Grid item={ true } xs={ 3 }>
                                                        <Typography gutterBottom={ true } variant="body2"
                                                                    color="textSecondary">
                                                            FPS
                                                        </Typography>
                                                        <Typography variant="h5">
                                                            { fps.toFixed(2) }
                                                        </Typography>
                                                    </Grid>
                                                    <Grid item={ true } xs={ 3 }>
                                                        <Typography gutterBottom={ true } variant="body2"
                                                                    color="textSecondary">
                                                            Memory
                                                        </Typography>
                                                        <Typography variant="h5" component="h5">
                                                            { (memory).toFixed(2) }MB
                                                        </Typography>
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
                                                <TableCell component="th" scope="row" style={ {whiteSpace: 'nowrap'} }>
                                                    { moment(entry.CreatedAt).format('YYYY/MM/DD HH:mm:ss') }
                                                </TableCell>
                                                <TableCell>{ entry.Level }</TableCell>
                                                <TableCell>{ entry.Log }</TableCell>
                                            </TableRow>) }
                                        </TableBody>
                                    </Table>
                                </TabPanel>
                                <TabPanel value={ value } index={ 2 }>
                                    <ProtocolLogComponent entries={ entries }/>
                                </TabPanel>
                                <TabPanel value={ value } index={ 3 }>
                                    <ProtocolScreensComponent entries={ screenEntries }/>
                                </TabPanel>
                                <TabPanel value={ value } index={ 4 }>
                                    Video (not implemented)
                                </TabPanel>
                                <TabPanel value={ value } index={ 5 }>
                                    <Grid container={ true } sx={{padding: 1}} spacing={1}>
                                        {checkpoints !== undefined && checkpoints.length > 0 && <Grid item={ true } xs={ 12 }>
                                            <Typography gutterBottom={ true } variant="subtitle1">
                                                Checkpoints
                                            </Typography>
                                            <Divider />
                                            <ResponsiveContainer width={ "100%" } height={ 200 }>
                                                <LineChart
                                                    width={ 600 }
                                                    height={ 300 }
                                                    data={ checkpoints }
                                                    margin={ {top: 10, right: 20, left: 10, bottom: 5} }
                                                >
                                                    <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3" />
                                                    <XAxis dataKey="Checkpoint"/>
                                                    <YAxis />
                                                    <Tooltip/>
                                                    <Legend />
                                                    <Line type="monotone" dataKey="Runtime" stroke="#ff7300">
                                                        <LabelList content={<CustomLineLabel unit={'s'}/>}/>
                                                    </Line>
                                                    <Line type="monotone" dataKey="ExecutionTime" stroke="#8884d8">
                                                        <LabelList content={<CustomLineLabel unit={'s'}/>}/>
                                                    </Line>
                                                </LineChart>
                                            </ResponsiveContainer>
                                        </Grid>}
                                        <Grid item={ true } xs={ 12 }>
                                            <Typography gutterBottom={ true } variant="subtitle1">
                                                FPS
                                            </Typography>
                                            <Divider />
                                            <ResponsiveContainer width={ "100%" } height={ 200 }>
                                                <LineChart
                                                    width={ 600 }
                                                    height={ 300 }
                                                    data={ performanceEntries }
                                                    syncId="anyId"
                                                    margin={ {top: 10, right: 20, left: 10, bottom: 5} }
                                                >
                                                    <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                                    <XAxis dataKey="Checkpoint"/>
                                                    <YAxis />
                                                    <Tooltip/>
                                                    <Line type="monotone" dataKey="FPS" stroke="#ff7300" yAxisId={ 0 }>
                                                        <LabelList content={<CustomLineLabel unit={'fsp'}/>}/>
                                                    </Line>
                                                </LineChart>
                                            </ResponsiveContainer>
                                        </Grid>
                                        <Grid item={ true } xs={ 12 }>
                                            <Typography gutterBottom={ true } variant="subtitle1">
                                                Memory
                                            </Typography>
                                            <Divider />
                                            <ResponsiveContainer width={ "100%" } height={ 200 }>
                                                <LineChart
                                                    width={ 600 }
                                                    height={ 300 }
                                                    data={ performanceEntries }
                                                    syncId="anyId"
                                                    margin={ {top: 10, right: 20, left: 10, bottom: 5} }
                                                >
                                                    <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                                    <XAxis dataKey="Checkpoint"/>
                                                    <YAxis />
                                                    <Tooltip/>
                                                    <Line type="monotone" dataKey="MEM" stroke="#ff7300" yAxisId={ 0 }>
                                                        <LabelList content={<CustomLineLabel unit={'MB'}/>}/>
                                                    </Line>
                                                </LineChart>
                                            </ResponsiveContainer>
                                        </Grid>
                                        { cpu > 0 && (
                                            <Grid item={ true } xs={ 12 }>
                                                <Typography gutterBottom={ true } variant="subtitle1">
                                                    CPU
                                                </Typography>
                                                <Divider />
                                                <ResponsiveContainer width={ "100%" } height={ 200 }>
                                                    <LineChart
                                                        width={ 600 }
                                                        height={ 300 }
                                                        data={ performanceEntries }
                                                        syncId="anyId"
                                                        margin={ {top: 10, right: 20, left: 10, bottom: 5} }
                                                    >
                                                        <CartesianGrid stroke="#f5f5f5" strokeDasharray="3 3"/>
                                                        <XAxis dataKey="Checkpoint"/>
                                                        <YAxis />
                                                        <Tooltip/>
                                                        <Line type="monotone" dataKey="CPU" stroke="#ff7300" yAxisId={ 0 }>
                                                            <LabelList content={<CustomLineLabel unit={'%'}/>}/>
                                                        </Line>
                                                    </LineChart>
                                                </ResponsiveContainer>
                                            </Grid>
                                        ) }
                                    </Grid>
                                </TabPanel>
                            </Grid>
                        </Paper>
                    </TitleCard>
                </Grid>
            </Grid>
        </Grid>
    );
};

export default TestProtocolContent;
