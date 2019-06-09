import React from 'react';
import {DataContext} from './ChatApp';

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
            ({userName, connected}) => (
                <div className='sidebar'>
                    <div className="card blue-grey darken-1">
                        <div className="card-content white-text">
                            <span className="card-title">{userName}</span>
                        </div>
                        <div className="card-action">
                            <a href="#0" className={'status'}>
                                {
                                    connected
                                        ? <i className='material-icons tiny'>check_circle</i>
                                        : <i className='material-icons tiny'>sync</i>
                                }
                                &nbsp;Connect{connected ? 'ed' : 'ing...'}</a>
                        </div>
                    </div>

                    <ul className="collection with-header">
                        <li className="collection-header"><h6>Channels</h6></li>
                        <li className="collection-item">Alvin</li>
                        <li className="collection-item">Alvin</li>
                        <li className="collection-item">Alvin</li>
                        <li className="collection-item">Alvin</li>
                    </ul>
                </div>
            )
        }
    </DataContext.Consumer>

);

export const ChatDialogue = () => {
    return <div className="dialogue">
        <div className={'chat'}>
            <div className={'messages'}>
                content

            </div>
            <div className={'text-input'}>
                <textarea></textarea>
                <button>Submit</button>
            </div>
        </div>
    </div>
};