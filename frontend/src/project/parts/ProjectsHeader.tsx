import React from 'react';
import {Box, Button, InputAdornment, Stack, TextField, Typography} from '@mui/material';
import {AddRounded, SearchRounded} from '@mui/icons-material';

export interface ProjectsHeaderProps {
    // the search box shows only for a long list
    searchable: boolean,
    search: string,
    onSearch: (search: string) => void,
    onNew: () => void,
}

const SearchBox: React.FC<Pick<ProjectsHeaderProps, 'search' | 'onSearch'>> = ({search, onSearch}) =>
    <TextField size="small" placeholder="Find a project or app" value={search} onChange={e => onSearch(e.target.value)}
        slotProps={{
                   input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                   htmlInput: {'aria-label': 'Find a project'},
               }}/>;

const ProjectsHeader: React.FC<ProjectsHeaderProps> = ({searchable, search, onSearch, onNew}) =>
    <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} useFlexGap={true} alignItems={{sm: 'center'}} sx={{mb: 3}}>
        <Box>
            <Typography variant="h5" component="h1">Projects</Typography>
            <Typography variant="body2" color="text.secondary">Pick a project to see its apps, tests and devices.</Typography>
        </Box>
        <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {sm: 'auto'}}}>
            {searchable && <SearchBox search={search} onSearch={onSearch}/>}
            <Button variant="contained" startIcon={<AddRounded/>} onClick={onNew}>New project</Button>
        </Stack>
    </Stack>;

export default ProjectsHeader;
