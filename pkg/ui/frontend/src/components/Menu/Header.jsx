import React from 'react';
import { connect } from 'react-redux';
import { Link } from 'react-router-dom';

import makeStyles from '@material-ui/styles/makeStyles';
import { withStyles } from '@material-ui/core/styles';

import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';

import Menu from './Menu';

import { toggleMenu } from '../../actions/generalActions';

const useStyles = makeStyles({
    menuButton: {
        marginLeft: -12,
        marginRight: 20,
    },
});

const Header = props => {
    const classes = useStyles();

    const toggleMenu = event => {
        props.toggleMenu(true);
    };

    return (
        <div>
            <AppBar position={'static'} color={'primary'}>
                {/*<Toolbar>*/}
                <MenuIcon />
                <Menu />
            </AppBar>
        </div>
    );
};

export default connect(null, { toggleMenu })(Header);
