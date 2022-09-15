import React, { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import TestsTable from '../../components/tests-table.component';
import { useNavigate } from 'react-router-dom';
import { Divider, FormControl, MenuItem, Select, SelectChangeEvent, Typography } from '@mui/material';
import { ApplicationProps } from "../../application/application.props";
import { useProjectContext } from "../../project/project.context";
import { useProjectAppContext } from "../../project/app.context";
import { TitleCard } from "../../components/title.card.component";

const Tests: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const { projectId } = useProjectContext();
    const { appId } = useProjectAppContext();

    const {appState, dispatch} = props;

    const navigate = useNavigate();

    function newTestClick(): void {
        navigate(`/project/${projectId}/app/${value}/test/new`)
    }

    const [value, setValue] = useState<string>(appId !== null ? `${ appId }` : appState.project?.Apps === undefined ? '' : appState.project?.Apps.length === 0 ? '' : `${ appState.project?.Apps[ 0 ].ID }`);
    const handleChange = (event: SelectChangeEvent): void => {
        setValue(event.target.value as string);
    };

    useEffect(() => {
        if (appId === 0) {
            const value = appState.project?.Apps === undefined ? null : appState.project?.Apps.length === 0 ? null : appState.project?.Apps[ 0 ].ID;
            if (value !== null) {
                navigate(`/project/${projectId}/app/${value}/tests`)
            }
        }
    }, [appState.project?.Apps])

    return (
        <Grid container={ true } spacing={ 2 }>
            <Grid item={ true } xs={ 12 }>
                <Typography variant={ "h1" }>Test Cases <FormControl variant="standard">
                    <Select
                        id="app-select"
                        label="App"
                        defaultValue={`${ value }`}
                        value={`${ value }`}
                        autoWidth={ true }
                        disableUnderline={ true }
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
                    style={ {maxWidth: 800} }
                >
                    <TitleCard title={ "Tests" }>
                        <Paper sx={ {margin: 'auto', overflow: 'hidden'} }>
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
                                    <TestsTable appId={+value} />
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