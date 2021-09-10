import { FC, ReactElement, SyntheticEvent, useEffect, useState } from 'react';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import TestRunDataService from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import ITestRunData from '../../types/test.run';
import { TestResultState } from '../../types/test.result.state.enum';
import ITestProtocolData from '../../types/test.protocol';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import { AvTimer, Cancel, CheckCircle, DateRange } from '@material-ui/icons';
import { pink } from '@material-ui/core/colors';
import { Box, Tab, Tabs } from '@material-ui/core';
import Table from '@material-ui/core/Table';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import TableBody from '@material-ui/core/TableBody';
import Moment from 'react-moment';

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
            maxWidth: 936,
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
        TestRunDataService.getLast(testId).then(response => {
            console.log(response.data);
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

    function getDuration(p: ITestProtocolData): string {
        return 'running';
        if (p.EndedAt !== null && p.EndedAt !== undefined) {
            const end = p?.EndedAt?.valueOf();
            const start = p?.StartedAt?.valueOf();
            const duration = end - start;
            console.log(start);
            console.log(end);
            console.log(duration);
            const date = new Date(duration);
            return date.toTimeString().split(' ')[0];
        }
        return 'running';
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
                            {protocol?.TestResult == TestResultState.TestResultSuccess ?
                                <CheckCircle className={classes.block} color="success"/> :
                                <Cancel className={classes.block} sx={{ color: pink[500] }}/>}
                        </Grid>
                        <Grid item={true}>
                            {protocol?.TestResult == TestResultState.TestResultSuccess ? 'Success' : 'Failed'}
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
                            <Tab label="Logs" {...a11yProps(1)} />
                            <Tab label="Screenshots" {...a11yProps(2)} />
                            <Tab label="Video" {...a11yProps(3)} />
                            <Tab label="Performance" {...a11yProps(4)} />
                        </Tabs>
                    </Box>
                </AppBar>
                <TabPanel value={value} index={0}>
                    Status
                </TabPanel>
                <TabPanel value={value} index={1}>
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
                                <TableCell component="th" scope="row">
                                    <Moment format="YYYY/MM/DD HH:mm:ss">{entry.CreatedAt}</Moment>
                                </TableCell>
                                <TableCell>{entry.Source}</TableCell>
                                <TableCell>{entry.Level}</TableCell>
                                <TableCell>{entry.Message}</TableCell>
                            </TableRow>)}
                        </TableBody>
                    </Table>
                </TabPanel>
                <TabPanel value={value} index={2}>
                    Screenshots
                </TabPanel>
                <TabPanel value={value} index={3}>
                    Video
                </TabPanel>
                <TabPanel value={value} index={4}>
                    Performance
                </TabPanel>
            </Box>
        </Paper>
    );
};

export default withStyles(styles)(TestProtocol);
