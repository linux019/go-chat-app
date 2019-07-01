import React from 'react';
import PropTypes from 'prop-types';

class UsersList extends React.Component {
    static propTypes = {
        getUsersList: PropTypes.func.isRequired,
        openChannel: PropTypes.func.isRequired,
        users: PropTypes.object.isRequired,
    };

    componentDidMount() {
        this.props.getUsersList();
    }

    render() {
        const {users, openChannel} = this.props;
        return (
            <ul className="collection with-header users">
                <li className="collection-header"><h6>Users</h6></li>
                {
                    Object.keys(users).map(
                        name =>
                            <li key={name}
                                onClick={e => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    openChannel(name);
                                }}
                                className="collection-item">
                                {
                                    users[name].online
                                        ? <i className="material-icons light-green-text">lens</i>
                                        : <i className="material-icons grey-text">radio_button_unchecked</i>
                                }
                                {name}
                            </li>
                    )
                }
            </ul>
        )
    }
}

export default UsersList;