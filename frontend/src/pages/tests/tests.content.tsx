import React, { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import TestsTable from '../../components/tests-table.component';
import { useNavigate } from 'react-router-dom';
import { Divider, FormControl, MenuItem, Select, SelectChangeEvent, Typography } from '@mui/material';
import { ApplicationProps } from "../../application/application.props";
import { useProjectAppContext } from "../../project/app.context";
import { TitleCard } from "../../components/title.card.component";

const Tests: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const { projectId, appId } = useProjectAppContext();

    const {appState, dispatch} = props;

    const navigate = useNavigate();

    function newTestClick(): void {
        navigate(`/project/${projectId}/app/${appId}/test/new`)
    }

    const handleChange = (event: SelectChangeEvent): void => {
        if (appState.appId != +event.target.value) {
            navigate(`/project/${projectId}/app/${event.target.value}/tests`)
        }
    };

    useEffect(() => {
        if (appId === 0) {
            const value = appState.project?.Apps === undefined ? null : appState.project?.Apps.length === 0 ? null : appState.project?.Apps[ 0 ].ID;
            if (value !== null) {
                navigate(`/project/${projectId}/app/${value}/tests`)
            }
        }
    }, [appState.project?.Apps, appId])

    return (
        <Grid container={ true } spacing={ 2 }>
            <Grid item={ true } xs={ 12 }>
                <Typography variant={ "h1" }>Test Cases <FormControl variant="standard">
                    <Select
                        id="app-select"
                        label="App"
                        defaultValue={`${ appId }`}
                        value={`${ appId }`}
                        disableUnderline={ true }
                        fullWidth={true}
                        sx={ {color: "black"} }
                        onChange={handleChange}
                    >
                        {
                            appState.project?.Apps.map(app => (
                                <MenuItem key={ `app_item_${ app.ID }` } value={ app.ID }>{ app.Name }</MenuItem>
                            ))
                        }
                    </Select>
                </FormControl></Typography>
            </Grid>
            <Grid item={ true } xs={ 12 }>
                <Divider/>
            </Grid>
            <Grid item={ true } container={ true } xs={ 12 } alignItems={ "center" } justifyContent={ "center" }>
                <Grid
                    item={ true }
                    xs={ 12 }
                >
                    <TitleCard title={ "Tests" }>
                        <Paper sx={ {width: '100%', margin: 'auto', overflow: 'hidden'} }>
                            <Grid container={ true }>
                                <Grid item={ true } xs={ 6 } container={ true } sx={ {
                                    padding: 2,
                                    backgroundColor: '#fafafa',
                                    borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                                } }>
                                    Tests
                                </Grid>
                                <Grid item={ true } xs={ 6 } container={ true } justifyContent={ "flex-end" } sx={ {
                                    padding: 1,
                                    backgroundColor: '#fafafa',
                                    borderBottom: '1px solid rgba(0, 0, 0, 0.12)'
                                } }>
                                    <Button variant={ "contained" } onClick={ newTestClick }>Add new Test</Button>
                                </Grid>
                                <Grid item={ true } xs={ 12 }>
                                    <TestsTable appState={appState} appId={appState.appId} />
                                </Grid>
                            </Grid>
                        </Paper>
                    </TitleCard>
                </Grid>

            </Grid>
        </Grid>
    );
};

export default Tests;