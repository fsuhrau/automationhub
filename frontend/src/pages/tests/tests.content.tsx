import { FC } from 'react';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import Button from '@material-ui/core/Button';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import SearchIcon from '@material-ui/icons/Search';
import TestsTable from '../../components/tests-table.component';
import { useHistory } from 'react-router-dom';
import { AppContext } from '../../context/app.context';
import { Typography } from '@material-ui/core';

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
    });

export type TestsProps = WithStyles<typeof styles>;

const Tests: FC<TestsProps> = (props) => {
    const { classes } = props;

    const history = useHistory();

    function newTestClick(): void {
        history.push('/web/test/new');
    }

    return (
        <AppContext.Provider value={ { title: 'Tests' } }>
            <Typography variant={ 'h4' }>
                Tests
            </Typography>
            <br/>
            <Paper className={ classes.paper }>
                <AppBar className={ classes.searchBar } position="static" color="default" elevation={ 0 }>
                    <Toolbar>
                        <Grid container={ true } spacing={ 2 } alignItems="center">
                            <Grid item={ true }>
                                <SearchIcon className={ classes.block } color="inherit"/>
                            </Grid>
                            <Grid item={ true } xs={ true }>
                            </Grid>
                            <Grid item={ true }>
                                <Button color="primary" variant="contained"
                                    onClick={ newTestClick }>
                                    New Test
                                </Button>
                            </Grid>
                        </Grid>
                    </Toolbar>
                </AppBar>
                <TestsTable/>
            </Paper>
        </AppContext.Provider>
    );
};

export default withStyles(styles)(Tests);
