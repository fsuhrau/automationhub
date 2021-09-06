import { FC, useEffect, useState } from 'react';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import TestRunDataService from '../../services/test.run.service';
import { useParams } from 'react-router-dom';
import {
    Button,
} from '@material-ui/core';
import ITestRunData from '../../types/test.run';
import { TestResultState } from '../../types/test.result.state.enum';
import ITestProtocolData from '../../types/test.protocol';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import SearchIcon from '@material-ui/icons/Search';
import TextField from '@material-ui/core/TextField';
import Tooltip from '@material-ui/core/Tooltip';
import IconButton from '@material-ui/core/IconButton';
import RefreshIcon from '@material-ui/icons/Refresh';
import Typography from '@material-ui/core/Typography';
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

    return (
        <Paper className={classes.paper}>
            <AppBar className={classes.searchBar} position="static" color="default" elevation={0}>
                <Toolbar>
                    <Grid container={true} spacing={2} alignItems="center">
                        <Grid item={true}>
                            <SearchIcon className={classes.block} color="inherit"/>
                        </Grid>
                        <Grid item={true} xs={true}>
                            <TextField
                                fullWidth={true}
                                placeholder="Search by email address, phone number, or user UID your mam"
                                InputProps={{
                                    disableUnderline: true,
                                    className: classes.searchInput,
                                }}
                            />
                        </Grid>
                        <Grid item={true}>
                            <Button variant="contained" color="primary" className={classes.addUser}>
                                Add user
                            </Button>
                            <Tooltip title="Reload">
                                <IconButton>
                                    <RefreshIcon className={classes.block} color="inherit"/>
                                </IconButton>
                            </Tooltip>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>

            <div className={classes.contentWrapper}>
                <Typography color="textSecondary" align="center">
                    No users for this project yet
                </Typography>
            </div>
        </Paper>
    );
};

export default withStyles(styles)(TestProtocol);
