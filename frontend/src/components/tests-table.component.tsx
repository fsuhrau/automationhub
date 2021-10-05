import { ChangeEvent, FC, useEffect, useState } from 'react';
import { createStyles, withStyles, WithStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import ITestData from '../types/test';
import { executeTest, getAllTests } from '../services/test.service';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Link,
    TextField, Typography,
} from '@material-ui/core';
import { PlayArrow } from '@material-ui/icons';
import AppSelection from './app-selection.component';
import IAppData from '../types/app';
import { useHistory } from 'react-router-dom';

const styles = (): ReturnType<typeof createStyles> =>
    createStyles({
        table: {
            minWidth: 650,
        },
    });

export type TestProps = WithStyles<typeof styles>;

const TestsTable: FC<TestProps> = (props) => {
    const { classes } = props;

    const history = useHistory();

    // dialog
    const [open, setOpen] = useState(false);
    const handleClickOpen = (): void => {
        setOpen(true);
    };
    const handleClose = (): void => {
        setOpen(false);
    };

    // test handling
    const [selectedTestID, setSelectedTestID] = useState<number>(0);
    const [selectedAppID, setSelectedAppID] = useState<number>(0);
    const [envParameter, setEnvParameter] = useState<string>('');
    const [tests, setTests] = useState<ITestData[]>([]);

    useEffect(() => {
        getAllTests().then(response => {
            console.log(response.data);
            setTests(response.data);
        }).catch(e => {
            console.log(e);
        });
    }, []);

    const typeString = (type: number): string => {
        switch (type) {
            case 0: return 'Unity';
            case 1: return 'Cocos';
            case 2: return 'Serenity';
            case 3: return 'Scenario';
        }
        return '';
    };

    const executionString = (type: number): string => {
        switch (type) {
            case 0: return 'Concurrent';
            case 1: return 'Simultaneously';
        }
        return '';
    };

    const onRunTest = (id: number, appid: number): void => {
        executeTest(id, appid, envParameter).then(response => {
            history.push(`/web/test/${id}/run/${response.data.ID}`);
        }).catch(error => {
            console.log(error);
        });
    };

    const getDevices = (test: ITestData): string => {
        if (test.TestConfig.AllDevices) {
            return 'all';
        }
        if (test.TestConfig.Devices !== null) {
            return test.TestConfig.Devices.length.toString();
        }
        return 'n/a';
    };

    const getTests = (test: ITestData): string => {
        if (test.TestConfig.Unity !== undefined && test.TestConfig.Unity !== null) {
            if (test.TestConfig.Unity?.RunAllTests) {
                return 'all';
            }
            if (test.TestConfig.Unity.UnityTestFunctions !== null) {
                return test.TestConfig.Unity.UnityTestFunctions.length.toString();
            }
        }

        return 'n/a';
    };

    const onAppSelectionChanged = (app: IAppData): void => {
        setSelectedAppID(app.ID);
    };

    const onEnvParamsChanged = (event: ChangeEvent<HTMLInputElement>): void => {
        setEnvParameter(event.target.value);
    };

    return (
        <div>
            <Dialog open={open} onClose={handleClose} aria-labelledby="form-dialog-title">
                <DialogTitle id="form-dialog-title">App Selection</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Select an existing App to execute the tests.<br />
                        Or Upload a new one.<br />
                        <br />
                    </DialogContentText>
                    <AppSelection upload={true} classes={classes} onSelectionChanged={onAppSelectionChanged} />
                    You can change parameters of your app by providing key value pairs in an environment like format:<br />
                    <br />
                    <Typography variant={'subtitle2'}>
                        server=http://localhost:8080<br />
                        user=autohub
                    </Typography>
                    <br />
                    <TextField
                        id="outlined-multiline-static"
                        label="Parameter"
                        fullWidth={ true }
                        multiline={true}
                        rows={4}
                        defaultValue=""
                        variant="outlined"
                        onChange={ onEnvParamsChanged }
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleClose} color="primary">
                        Cancel
                    </Button>
                    <Button onClick={() => {
                        onRunTest(selectedTestID, selectedAppID);
                        handleClose();
                    }} color="primary" variant={'contained'} disabled={ selectedAppID === 0 }>
                        Start
                    </Button>
                </DialogActions>
            </Dialog>
            <TableContainer component={Paper}>
                <Table className={classes.table} size="small" aria-label="a dense table">
                    <TableHead>
                        <TableRow>
                            <TableCell>Name</TableCell>
                            <TableCell align="right">Typ</TableCell>
                            <TableCell align="right">Execution</TableCell>
                            <TableCell align="right">Devices</TableCell>
                            <TableCell align="right">Tests</TableCell>
                            <TableCell align="right"/>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {tests.map((test) => <TableRow key={test.Name}>
                            <TableCell component="th" scope="row">
                                <Link href={`test/${test.ID}/runs/last` } underline="none">
                                    {test.Name}
                                </Link>
                            </TableCell>
                            <TableCell align="right">{typeString(test.TestConfig.Type)}</TableCell>
                            <TableCell align="right">{executionString(test.TestConfig.ExecutionType)}</TableCell>
                            <TableCell align="right">{getDevices(test)}</TableCell>
                            <TableCell align="right">{getTests(test)}</TableCell>
                            <TableCell align="right">
                                <Button variant="outlined" color="primary" size="small" endIcon={<PlayArrow />} onClick={ () => {
                                    setSelectedTestID(test.ID as number);
                                    handleClickOpen();
                                }}>
                                    Run
                                </Button>
                            </TableCell>
                        </TableRow>)}
                    </TableBody>
                </Table>
            </TableContainer>
        </div>
    );
};

export default withStyles(styles)(TestsTable);
