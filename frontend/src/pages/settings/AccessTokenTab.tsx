import React, {useEffect, useState} from "react";
import {useHubState} from "../../hooks/HubStateProvider";
import {
    createAccessToken,
    deleteAccessToken,
    getAccessTokens,
    NewAccessTokenRequest
} from "../../services/settings.service";
import {useProjectContext} from "../../hooks/ProjectProvider";
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
import Moment from "react-moment";
import DeleteForeverIcon from "@mui/icons-material/DeleteForever";
import {LocalizationProvider, MobileDatePicker} from "@mui/x-date-pickers";
import {AdapterDayjs} from "@mui/x-date-pickers/AdapterDayjs";
import {Dayjs} from "dayjs";
import Button from "@mui/material/Button";
import {TitleCard} from "../../components/title.card.component";
import {HubStateActions} from "../../application/HubState";
import {useError} from "../../ErrorProvider";

const AccessTokenTab: React.FC = () => {

    const {state, dispatch} = useHubState()
    const {projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const [newToken, setNewToken] = useState<NewAccessTokenRequest>({
        Name: '',
        ExpiresAt: null,
    });

    const handleDeleteAccessToken = (accessTokenID: number): void => {
        deleteAccessToken(projectIdentifier, accessTokenID).then(value => {
            dispatch({
                type: HubStateActions.AccessTokenDelete,
                payload: accessTokenID,
            })
        }).catch(ex => {
            setError(ex)
        });
    };

    const createNewAccessToken = (): void => {
        createAccessToken(projectIdentifier, newToken).then(response => {
            setNewToken({
                Name: '',
                ExpiresAt: null,
            })

            dispatch({
                type: HubStateActions.AccessTokenAdd,
                payload: response.data
            })
        }).catch(ex => {
            setError(ex)
        });
    };

    useEffect(() => {
        getAccessTokens(projectIdentifier).then(response => {
            dispatch({
                type: HubStateActions.AccessTokensUpdate,
                payload: response.data
            })
        }).catch(ex => {
            setError(ex)
        });
    }, [projectIdentifier]);

    return (
        <TitleCard title={"Access Tokens"}>
            <Paper sx={{width: '100%', margin: 'auto', overflow: 'hidden'}}>
                <Box sx={{width: '100%'}}>
                    <Box sx={{p: 3}}>
                        <Table size="small" aria-label="a dense table">
                            <TableHead>
                                <TableRow>
                                    <TableCell>Name</TableCell>
                                    <TableCell align="right">Token</TableCell>
                                    <TableCell align="right">Expires</TableCell>
                                    <TableCell></TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {state.accessTokens?.map((accessToken) => <TableRow key={accessToken.ID}>
                                    <TableCell>{accessToken.Name}</TableCell>
                                    <TableCell>
                                        <Grid container={true} direction={"row"} spacing={1} justifyContent={"right"}
                                              alignItems={"center"}>
                                            <Grid>
                                                {accessToken.Token}
                                            </Grid>
                                            <Grid>
                                                <CopyToClipboard>
                                                    {({copy}) => (
                                                        <IconButton
                                                            color={'primary'}
                                                            size={'small'}
                                                            onClick={() => copy(accessToken.Token)}
                                                        >
                                                            <ContentCopy/>
                                                        </IconButton>
                                                    )}
                                                </CopyToClipboard>

                                            </Grid>
                                        </Grid>
                                    </TableCell>
                                    <TableCell align="right">{accessToken.ExpiresAt !== null ? (
                                        <Moment
                                            format="YYYY/MM/DD HH:mm:ss">{accessToken.ExpiresAt}</Moment>) : ('Unlimited')}</TableCell>
                                    <TableCell>
                                        <IconButton color="secondary" size="small" onClick={() => {
                                            handleDeleteAccessToken(accessToken.ID as number);
                                        }}>
                                            <DeleteForeverIcon/>
                                        </IconButton>
                                    </TableCell>
                                </TableRow>)}

                                <TableRow>
                                    <TableCell>
                                        <TextField id="new_name" placeholder="Name"
                                                   variant="outlined"
                                                   fullWidth={true}
                                                   value={newToken.Name} size="small"
                                                   onChange={event => setNewToken(prevState => ({
                                                       ...prevState,
                                                       Name: event.target.value
                                                   }))}/>
                                    </TableCell>
                                    <TableCell></TableCell>
                                    <TableCell align="right">
                                        <LocalizationProvider dateAdapter={AdapterDayjs}>
                                            <MobileDatePicker
                                                value={newToken.ExpiresAt}
                                                onChange={(newValue: Dayjs | null) => {
                                                    setNewToken(prevState => ({
                                                        ...prevState,
                                                        ExpiresAt: newValue
                                                    }))
                                                }}
                                            />
                                        </LocalizationProvider>
                                    </TableCell>
                                    <TableCell>
                                        <Button variant="contained" color="primary" size="small"
                                                onClick={createNewAccessToken}>
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

export default AccessTokenTab;