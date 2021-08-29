import React from 'react';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import {createStyles, Theme, withStyles, WithStyles} from '@material-ui/core/styles';
import SearchIcon from '@material-ui/icons/Search';
import TestsTable from "../../components/tests-table-component";
import Typography from "@material-ui/core/Typography";
import {Android} from "@material-ui/icons";

const styles = (theme: Theme) =>
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
    });

export interface TestsProps extends WithStyles<typeof styles> {
}

function Tests(props: TestsProps) {
    const {classes} = props;

    return (
        <Paper className={classes.paper}>
            <AppBar className={classes.searchBar} position="static" color="default" elevation={0}>
                <Toolbar>
                    <Grid container spacing={2} alignItems="center">
                        <Grid item>
                            <SearchIcon className={classes.block} color="inherit"/>
                        </Grid>
                        <Grid item xs>
                        </Grid>
                        <Grid item>
                            <Button variant="contained" color="primary" className={classes.addUser} onClick={() => console.log('/the/path') }>
                                New Test
                            </Button>
                        </Grid>
                    </Grid>
                </Toolbar>
            </AppBar>
            <TestsTable>
            </TestsTable>
        </Paper>
    );
}

export default withStyles(styles)(Tests);