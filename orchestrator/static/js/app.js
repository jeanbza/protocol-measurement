// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import '../css/app.scss' // required for wepback to build css

import React from 'react'
import {render} from 'react-dom'

import Sidebar from './components/Sidebar'
import ResultsArea from './components/ResultsArea'

class App extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            selectedRun: undefined,
        }
    }

    onRunChangeHandler(newRun) {
        this.setState({
            selectedRun: newRun,
        })
    }

    render() {
        const {selectedRun} = this.state

        return <div className="page">
            <Sidebar selectedRun={selectedRun} onRunChange={this.onRunChangeHandler.bind(this)}/>
            <ResultsArea selectedRun={selectedRun}/>
        </div>
    }
}

render(<App/>, document.getElementById('root'))