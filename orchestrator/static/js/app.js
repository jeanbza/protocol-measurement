import '../css/app.scss' // required for wepback to build css

import React from 'react'
import {render} from 'react-dom'

import Sidebar from './components/Sidebar'
import ResultsArea from './components/ResultsArea'

class App extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            selectedSet: undefined,
        }
    }

    onSetChangeHandler(newSet) {
        this.setState({
            selectedSet: newSet,
        })
    }

    render() {
        const {selectedSet} = this.state

        return <div className="page">
            <Sidebar selectedSet={selectedSet} onSetChange={this.onSetChangeHandler.bind(this)}/>
            <ResultsArea selectedSet={selectedSet}/>
        </div>
    }
}

render(<App/>, document.getElementById('root'))