import React, {useEffect, useState} from "react";
import {useHubState} from "../../hooks/HubStateProvider";
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import Paper from "@mui/material/Paper";
import {Box} from "@mui/material";
import Table from "@mui/material/Table";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import TableBody from "@mui/material/TableBody";
import {getUsers} from "../../services/settings.service";
import {HubStateActions} from "../../application/HubState";
import {useError} from "../../ErrorProvider";

const UsersTab: React.FC = () => {

    const {state, dispatch} = useHubState()
    const {projectIdentifier} = useProjectContext();

    const {setError} = useError()

    useEffect(() => {
        getUsers(projectIdentifier).then(users => {
            dispatch({
                type: HubStateActions.UsersUpdate,
                payload: users
            })
        }).catch(e => {
            setError(e)
        });
    }, [projectIdentifier]);

    return (
        <TitleCard title={"Users and Permission"}>
            <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
                <Box sx={{width: '100%'}}>
                    <Box sx={{p: 3}}>
                        <Table size="small" aria-label="a dense table">
                            <TableHead>
                                <TableRow>
                                    <TableCell>Name</TableCell>
                                    <TableCell align="right">Email</TableCell>
                                    <TableCell align="right">role</TableCell>
                                    <TableCell></TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {state.users?.map((user) => <TableRow key={`user_${user.id}`}>
                                    <TableCell>{user.name}</TableCell>
                                    <TableCell>{user.email}</TableCell>
                                    <TableCell>{user.role}</TableCell>
                                    <TableCell>
                                    </TableCell>
                                </TableRow>)}
                            </TableBody>
                        </Table>
                    </Box>
                </Box>
            </Paper>
        </TitleCard>

    )
}

export default UsersTab;