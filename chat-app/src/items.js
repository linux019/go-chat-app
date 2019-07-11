import React from 'react';
import {DataContext} from './ChatApp';
import classnames from 'classnames';
import UsersList from './UsersList';

const Names = [
    '',
    'Teddie Miller',
    'Reece Sharp',
    'Hamish Matthews',
    'Charles Burns',
    'Eric Palmer',
    'Caelan Green',
    'Wilfred Khan',
    'Alex Allen',
    'Charlie Stevens',
    'Harley Robertson'
];

export const WelcomeScreen = ({name, setData}) => (
    <div className='container valign-wrapper'>
        <GitHubLink/>
        <div className='row'>
            <h6 className='center'>Select your name to enter a chat</h6>
            <div className={'input-field  col s12 m12'}>
                <select value={name} onChange={e => setData({name: e.target.value})}>
                    {
                        Names.map(item => <option key={item} value={item}>{item || 'Pick a name'}</option>)
                    }
                </select>
            </div>
            <input type={'text'}
                   value={name}
                   placeholder='Name'
                   onChange={e => setData({name: e.target.value})}
            />
            <button className="btn waves-effect waves-light right"
                    disabled={!name}
                    type="submit"
                    onClick={e => {
                        e.preventDefault();
                        setData({name, openChat: true});
                    }}
                    name="action">
                Submit
                <i className="material-icons right">send</i>
            </button>
        </div>
    </div>
);


export const Sidebar = () => (
    <DataContext.Consumer>
        {
            ({connected, getUsersList, askForChannelName, users, setActiveChannel}) => (
                <div className='sidebar'>
                    <GitHubLink/>
                    <UserNameCard/>

                    {
                        connected &&
                        <a className="waves-effect waves-light btn-small new-channel"
                           href='#0'
                           onClick={askForChannelName}>
                            <i className="material-icons left">control_point</i>
                            Create channel
                        </a>
                    }

                    <ChannelsList/>

                    {
                        connected &&
                        <UsersList {...{getUsersList, users, setActiveChannel}}/>
                    }
                </div>
            )
        }
    </DataContext.Consumer>
);

const UserNameCard = () => (
    <DataContext.Consumer>
        {
            ({connected, userName}) =>
                <div className="card blue-grey darken-1">
                    <div className="card-content white-text">
                        <span className="card-title">{userName}</span>
                    </div>
                    <div className="card-action">
                        <a href="#0" className='status'>
                            <i className='material-icons tiny'>{connected ? 'check_circle' : 'sync'}</i>
                            &nbsp;{connected ? 'Connected' : 'Connecting...'}</a>
                    </div>
                </div>
        }
    </DataContext.Consumer>
);

const ChannelsList = () => (
    <DataContext.Consumer>
        {
            ({channels, unreadChannels, activeChannel, setActiveChannel}) =>
                <ul className="collection with-header">
                    <li className="collection-header"><h6>Channels</h6></li>
                    {
                        Object.keys(channels).map(
                            ch =>
                                <li key={ch}
                                    onClick={e => {
                                        e.preventDefault();
                                        e.stopPropagation();
                                        setActiveChannel(ch, false);
                                    }}
                                    className={classnames("collection-item", activeChannel && activeChannel.name === ch && 'active')}>
                                    {
                                        unreadChannels[ch] &&
                                        <i className="material-icons left tiny light-green-text message">message</i>
                                    }
                                    {ch}
                                </li>
                        )
                    }
                </ul>
        }
    </DataContext.Consumer>
);


export const GitHubLink = () => (
    <a href="https://github.com/devbazilio/go-chat-app"
       target="_blank"
       className='repo-logo'
       rel="noopener noreferrer">
        <img alt="Link to Github repo"
             src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB2ZXJzaW9uPSIxLjEiIGlkPSJMYXllcl8xIiB4PSIwcHgiIHk9IjBweCIgd2lkdGg9IjQwcHgiIGhlaWdodD0iNDBweCIgdmlld0JveD0iMTIgMTIgNDAgNDAiIGVuYWJsZS1iYWNrZ3JvdW5kPSJuZXcgMTIgMTIgNDAgNDAiIHhtbDpzcGFjZT0icHJlc2VydmUiPjxwYXRoIGZpbGw9IiMzMzMzMzMiIGQ9Ik0zMiAxMy40Yy0xMC41IDAtMTkgOC41LTE5IDE5YzAgOC40IDUuNSAxNS41IDEzIDE4YzEgMC4yIDEuMy0wLjQgMS4zLTAuOWMwLTAuNSAwLTEuNyAwLTMuMiBjLTUuMyAxLjEtNi40LTIuNi02LjQtMi42QzIwIDQxLjYgMTguOCA0MSAxOC44IDQxYy0xLjctMS4yIDAuMS0xLjEgMC4xLTEuMWMxLjkgMC4xIDIuOSAyIDIuOSAyYzEuNyAyLjkgNC41IDIuMSA1LjUgMS42IGMwLjItMS4yIDAuNy0yLjEgMS4yLTIuNmMtNC4yLTAuNS04LjctMi4xLTguNy05LjRjMC0yLjEgMC43LTMuNyAyLTUuMWMtMC4yLTAuNS0wLjgtMi40IDAuMi01YzAgMCAxLjYtMC41IDUuMiAyIGMxLjUtMC40IDMuMS0wLjcgNC44LTAuN2MxLjYgMCAzLjMgMC4yIDQuNyAwLjdjMy42LTIuNCA1LjItMiA1LjItMmMxIDIuNiAwLjQgNC42IDAuMiA1YzEuMiAxLjMgMiAzIDIgNS4xYzAgNy4zLTQuNSA4LjktOC43IDkuNCBjMC43IDAuNiAxLjMgMS43IDEuMyAzLjVjMCAyLjYgMCA0LjYgMCA1LjJjMCAwLjUgMC40IDEuMSAxLjMgMC45YzcuNS0yLjYgMTMtOS43IDEzLTE4LjFDNTEgMjEuOSA0Mi41IDEzLjQgMzIgMTMuNHoiLz48L3N2Zz4="/>
    </a>
);