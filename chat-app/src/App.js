import React from 'react';
import './App.scss';
// Teddie Miller
// Reece Sharp
// Hamish Matthews
// Charles Burns
// Eric Palmer
// Caelan Green
// Wilfred Khan
// Alex Allen
// Charlie Stevens
// Harley Robertson
function App() {
    return (
        <>
            <div className='sidebar'>
                <div className="card blue-grey darken-1">
                    <div className="card-content white-text">
                        <span className="card-title">Card Title</span>
                        <p>I am a very simple card</p>
                    </div>
                    <div className="card-action">
                        <a href="#" className={'status'}>
                            <i className={'material-icons tiny'}>check_circle</i>
                            &nbsp;Connected</a>
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
            <div className="dialogue">
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
        </>
    );
}

export default App;
