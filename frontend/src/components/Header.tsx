import * as React from 'react';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';
import NavbarBreadcrumbs from './NavbarBreadcrumbs';
import ColorModeIconDropdown from '../shared-theme/ColorModeIconDropdown';

// Header sits above every project page and says where you are.
export default function Header() {
  return (
    <Stack
      direction="row"
      sx={{
        width: '100%',
        alignItems: 'center',
        justifyContent: 'space-between',
        maxWidth: { sm: '100%', md: '1700px' },
        pt: 1.5,
      }}
      spacing={2}
    >
      <Box sx={{ minWidth: 0 }}>
        <NavbarBreadcrumbs />
      </Box>
      <Box sx={{ display: { xs: 'none', md: 'block' } }}>
        <ColorModeIconDropdown />
      </Box>
    </Stack>
  );
}
