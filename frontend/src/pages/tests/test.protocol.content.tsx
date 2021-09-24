import { FC, ReactElement, SyntheticEvent, useEffect, useState } from 'react';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import { getLastRun } from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import { TestResultState } from '../../types/test.result.state.enum';
import ITestProtocolData from '../../types/test.protocol';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import { AvTimer, DateRange } from '@material-ui/icons';
import { Box, Button, Tab, Tabs } from '@material-ui/core';
import Table from '@material-ui/core/Table';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import TableBody from '@material-ui/core/TableBody';
import Moment from 'react-moment';
import moment from 'moment';
import IProtocolEntryData from '../../types/protocol.entry';
import TestStatusIconComponent from '../../components/test-status-icon.component';
import TestStatusTextComponent from '../../components/test-status-text.component';

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
            hidden={value !== index}
            id={`simple-tabpanel-${index}`}
            aria-labelledby={`simple-tab-${index}`}
            {...other}
        >
            {value === index && (
                <Box sx={{ p: 3 }}>
                    <Typography>{children}</Typography>
                </Box>
            )}
        </div>
    );
}


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
        block: {
            display: 'block',
        },
        addUser: {
            marginRight: theme.spacing(1),
        },
        contentWrapper: {
            margin: '40px 16px',
        },
        table: {
            minWidth: 650,
        },
    });

function a11yProps(index: number): Map<string, string> {
    return {
        id: `simple-tab-${index}`,
        'aria-controls': `simple-tabpanel-${index}`,
    };
}

type TestProtocolProps = WithStyles<typeof styles>;

const TestProtocol: FC<TestProtocolProps> = (props) => {
    const { classes } = props;

    const { testId } = useParams<number>();
    const { protocolId } = useParams<number>();

    const [run, setRun] = useState<ITestRunData>();
    const [protocol, setProtocol] = useState<ITestProtocolData>();


    useEffect(() => {
        getLastRun(testId).then(response => {
            setRun(response.data);
            for (let i = 0; i < response.data.Protocols.length; ++i) {
                if (response.data.Protocols[i].ID == +protocolId) {
                    setProtocol(response.data.Protocols[i]);
                    break;
                }
            }
        }).catch(ex => {
            console.log(ex);
        });
    }, [testId, protocolId]);

    function getDuration(p: ITestProtocolData | undefined): string {
        if (p === undefined) {
            return '';
        }
        if (p.EndedAt !== null && p.EndedAt !== undefined) {
            const duration = (new Date(p.EndedAt)).valueOf() - (new Date(p.StartedAt)).valueOf();
            const m = moment.utc(duration);
            const secs = duration / 1000;
            if (secs > 60 * 60) {
                return m.format('h') + 'Std ' + m.format('m') + 'Min ' + m.format('s') + 'Sec';
            }
            if (secs > 60) {
                return m.format('m') + 'Min ' + m.format('s') + 'Sec';
            }
            return m.format('s') + 'Sec';
        }
        return 'running';
    }

    function lastEntries(): Array<IProtocolEntryData> {
        const lastE = Array<IProtocolEntryData>();
        const length = protocol?.Entries?.length;
        for (let i = length - 1; i > Math.max(length - 3, 0); i--) {
            lastE.push(protocol?.Entries[i]);
        }
        return lastE;
    }

    const [value, setValue] = useState(0);

    const handleChange = (event: SyntheticEvent, newValue: number): void => {
        setValue(newValue);
    };

    return (
        <Paper className={classes.paper}>
            <AppBar className={classes.searchBar} position="static" color="default" elevation={0}>
                <Toolbar>
                    <Grid container={true} spacing={2} alignItems="center">
                        <Grid item={true}>
                            <TestStatusIconComponent status={protocol?.TestResult} />
                        </Grid>
                        <Grid item={true}>
                            <TestStatusTextComponent status={protocol?.TestResult} />
                        </Grid>
                        <Grid item={true}>
                            <DateRange className={classes.block} color="inherit"/>
                        </Grid>
                        <Grid item={true}>
                            <Moment format="YYYY/MM/DD HH:mm:ss">{protocol?.StartedAt}</Moment>
                        </Grid>

                        <Grid item={true}>
                            <AvTimer className={classes.block} color="inherit"/>
                        </Grid>
                        <Grid item={true} xs={true}>
                            {getDuration(protocol)}
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <Box sx={{ width: '100%' }}>
                <AppBar className={classes.searchBar} position="static" color="default" elevation={0}>
                    <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                        <Tabs
                            value={value}
                            onChange={handleChange}
                        >
                            <Tab label="Status" {...a11yProps(0)} />
                            <Tab label="Executer" {...a11yProps(1)} />
                            <Tab label="Logs" {...a11yProps(2)} />
                            <Tab label="Screenshots" {...a11yProps(3)} />
                            <Tab label="Video/Replay" {...a11yProps(4)} />
                            <Tab label="Performance" {...a11yProps(5)} />
                        </Tabs>
                    </Box>
                </AppBar>
                <TabPanel value={value} index={0}>
                    <Grid container={true} direction="column" spacing={6}>
                        <Grid item={true}>
                            <Grid container={true} spacing={6}>
                                {protocol?.TestResult == TestResultState.TestResultFailed &&
                                <Box width='100%' height='100%' color="white" bgcolor="palevioletred" padding={2}>
                                    <Grid container={true} spacing={6} justifyContent="center"
                                        alignItems="center">
                                        <Grid item={true}>
                                            {lastEntries().map((data) => <Typography>
                                                <Moment
                                                    format="YYYY/MM/DD HH:mm:ss">{data.CreatedAt}</Moment>: {data.Message}
                                            </Typography>,
                                            )}
                                        </Grid>
                                        <Grid item={true} sx={2}>
                                            <Button variant="contained">Show</Button>
                                        </Grid>
                                    </Grid>
                                </Box>
                                }
                            </Grid>
                        </Grid>
                        <Grid item={true}>
                            <Grid container={true} spacing={6}>
                                <Grid item={true}>
                                    <Grid item={true} xs={true} container={true} direction="column" spacing={4}>
                                        <Grid item={true} xs={true}>
                                            <Typography gutterBottom={true} variant="subtitle1">
                                                Time
                                            </Typography>
                                            <Typography gutterBottom={true} variant="body2" color="textSecondary">
                                                Execution
                                            </Typography>
                                            <Typography variant="h5" component="h5" style={{ whiteSpace: 'nowrap' }}>
                                                {getDuration(protocol)}
                                            </Typography>
                                        </Grid>
                                    </Grid>
                                </Grid>
                                <Grid item={true} xs={12} sm={true} container={true}>
                                    <Grid item={true} xs={true} container={true} direction="column" spacing={2}>
                                        <Grid item={true} xs={true}>
                                            <Typography gutterBottom={true} variant="subtitle1">
                                                Details
                                            </Typography>
                                            <Typography gutterBottom={true} variant="gutterBottom">
                                                <Grid container={true} spacing={2}>
                                                    <Grid item={true} xs={2}>
                                                        <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                            Actions
                                                        </Typography>
                                                        <Typography variant="h5" component="h5">
                                                            {protocol?.Entries.length}
                                                        </Typography>
                                                    </Grid>
                                                    <Grid item={true} xs={2}>
                                                        <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                            StartupTime
                                                        </Typography>
                                                        <Typography variant="h5" component="h5">
                                                            10 sec
                                                        </Typography>
                                                    </Grid>
                                                    <Grid item={true} xs={2}>
                                                        <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                            FPS
                                                        </Typography>
                                                        <Typography variant="h5" component="h5">
                                                            12
                                                        </Typography>
                                                    </Grid>
                                                    <Grid item={true} xs={2}>
                                                        <Typography gutterBottom={true} variant="body2"
                                                            color="textSecondary">
                                                            Memory
                                                        </Typography>
                                                        <Typography variant="h5" component="h5">
                                                            400MB
                                                        </Typography>
                                                    </Grid>
                                                </Grid>
                                            </Typography>
                                        </Grid>
                                    </Grid>
                                </Grid>
                            </Grid>
                        </Grid>
                    </Grid>
                </TabPanel>
                <TabPanel value={value} index={1}>
                    <Table className={classes.table} size="small" aria-label="a dense table">
                        <TableHead>
                            <TableRow>
                                <TableCell>Date</TableCell>
                                <TableCell>Level</TableCell>
                                <TableCell>Log</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {run?.Log?.map((entry) => <TableRow key={entry.ID}>
                                <TableCell component="th" scope="row" style={{ whiteSpace: 'nowrap' }}>
                                    <Moment format="YYYY/MM/DD HH:mm:ss">{entry.CreatedAt}</Moment>
                                </TableCell>
                                <TableCell>{entry.Level}</TableCell>
                                <TableCell>{entry.Log}</TableCell>
                            </TableRow>)}
                        </TableBody>
                    </Table>
                </TabPanel>
                <TabPanel value={value} index={2}>
                    missing filter options
                    <Table className={classes.table} size="small" aria-label="a dense table">
                        <TableHead>
                            <TableRow>
                                <TableCell>Date</TableCell>
                                <TableCell>Source</TableCell>
                                <TableCell>Level</TableCell>
                                <TableCell>Info</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {protocol?.Entries?.map((entry) => <TableRow key={entry.ID}>
                                <TableCell component="th" scope="row" style={{ whiteSpace: 'nowrap' }}>
                                    <Moment format="YYYY/MM/DD HH:mm:ss">{entry.CreatedAt}</Moment>
                                </TableCell>
                                <TableCell>{entry.Source}</TableCell>
                                <TableCell>{entry.Level}</TableCell>
                                <TableCell>{entry.Message}</TableCell>
                            </TableRow>)}
                        </TableBody>
                    </Table>
                </TabPanel>
                <TabPanel value={value} index={3}>
                    screenshots?
                </TabPanel>
                <TabPanel value={value} index={4}>
                    Video
                </TabPanel>
                <TabPanel value={value} index={5}>
                    Performance
                </TabPanel>
            </Box>
        </Paper>
    );
};

export default withStyles(styles)(TestProtocol);
