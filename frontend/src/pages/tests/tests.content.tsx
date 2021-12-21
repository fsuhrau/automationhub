import React from 'react';

import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Paper from '@mui/material/Paper';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import TestsTable from '../../components/tests-table.component';
import { useHistory } from 'react-router-dom';
import { AppContext } from '../../context/app.context';
import { Typography } from '@mui/material';

const Tests: React.FC = () => {

    const history = useHistory();

    function newTestClick(): void {
        history.push('/web/test/new');
    }

    return (
        <AppContext.Provider value={ { title: 'Tests' } }>
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
                                <Typography variant={ 'h6' }>
                                    Tests
                                </Typography>
                            </Grid>
                            <Grid item={ true } xs={ true }>
                            </Grid>
                            <Grid item={ true }>
                                <Button variant="text" color="primary" size="small"
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

export default Tests;