import React, {useEffect, useState} from "react";
import {useHubState} from "../../hooks/HubStateProvider";
import {useProjectContext} from "../../hooks/ProjectProvider";
import {TitleCard} from "../../components/title.card.component";
import Paper from "@mui/material/Paper";
import {Box, TextField} from "@mui/material";
import Table from "@mui/material/Table";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TableCell from "@mui/material/TableCell";
import TableBody from "@mui/material/TableBody";
import Grid from "@mui/material/Grid";
import CopyToClipboard from "../../components/copy.clipboard.component";
import IconButton from "@mui/material/IconButton";
import {ContentCopy} from "@mui/icons-material";
import DeleteForeverIcon from "@mui/icons-material/DeleteForever";
import Button from "@mui/material/Button";
import {INodeData} from "../../types/node";
import {createNode, deleteNode, getNodes, NewNodeRequest} from "../../services/settings.service";
import {HubStateActions} from "../../application/HubState";
import {useError} from "../../ErrorProvider";

const NodesTab: React.FC = () => {

    const {state, dispatch} = useHubState()
    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const [newNode, setNewNode] = useState<NewNodeRequest>({
        name: '',
    });

    const handleCreateNode = (): void => {
        createNode(projectIdentifier, newNode).then(node => {
            setNewNode({
                name: '',
            })

            dispatch({
                type: HubStateActions.NodeAdd,
                payload: node
            })
        }).catch(exx => {
            setError(exx);
        });
    };


    const handleDeleteNode = (nodeID: number): void => {
        deleteNode(projectIdentifier, nodeID).then(value => {
            dispatch({
                type: HubStateActions.NodeDelete,
                payload: nodeID,
            })
        }).catch(ex => {
            setError(ex)
        });
    };

    useEffect(() => {
        getNodes(projectIdentifier).then(nodes => {
            dispatch({
                type: HubStateActions.NodesUpdate,
                payload: nodes
            })
        }).catch(ex => {
            setError(ex);
        });
    }, [projectIdentifier]);

    return (
        <TitleCard title={"Nodes"}>
            <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
                <Box sx={{width: '100%'}}>
                    <Box sx={{p: 3}}>
                        <Table size="small" aria-label="a dense table">
                            <TableHead>
                                <TableRow>
                                    <TableCell>Name</TableCell>
                                    <TableCell align="right">Identifier</TableCell>
                                    <TableCell></TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {state.nodes?.map((node) => <TableRow key={node.id}>
                                    <TableCell>{node.name}</TableCell>
                                    <TableCell>
                                        <Grid container={true} direction={"row"} spacing={1} justifyContent={"right"} alignItems={"center"}>
                                            <Grid>
                                                {node.identifier}
                                            </Grid>
                                            <Grid>
                                                <CopyToClipboard>
                                                    {({copy}) => (
                                                        <IconButton
                                                            color={'primary'}
                                                            size={'small'}
                                                            onClick={() => copy(node.identifier)}
                                                        >
                                                            <ContentCopy/>
                                                        </IconButton>
                                                    )}
                                                </CopyToClipboard>

                                            </Grid>
                                        </Grid>
                                    </TableCell>
                                    <TableCell>
                                        <IconButton color="secondary" size="small" onClick={() => {
                                            handleDeleteNode(node.id as number);
                                        }}>
                                            <DeleteForeverIcon/>
                                        </IconButton>
                                    </TableCell>
                                </TableRow>)}

                                <TableRow>
                                    <TableCell>
                                        <TextField id="new_node_name" placeholder="Name"
                                                   variant="outlined"
                                                   fullWidth={true}
                                                   value={newNode.name} size="small"
                                                   onChange={event => setNewNode(prevState => ({
                                                       ...prevState,
                                                       name: event.target.value
                                                   }))}/>
                                    </TableCell>
                                    <TableCell></TableCell>
                                    <TableCell>
                                        <Button variant="contained" color="primary" size="small"
                                                onClick={handleCreateNode}>
                                            Add
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            </TableBody>
                        </Table>
                    </Box>
                </Box>
            </Paper>
        </TitleCard>

    )
}

export default NodesTab;