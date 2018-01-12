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

import React from 'react'

class RunResults extends React.Component {
    constructor(props) {
        super(props)

        const {selectedRun, runTotalMessages} = props

        const runInterval = setInterval(() => fetch(new Request(`/runs/${selectedRun}`))
            .then(resp => resp.json())
            .then(run => this.setState({totalMessages: run.totalMessages, finishedCreating: run.finishedCreating}))
            .catch(err => {
                console.error(err)
                clearInterval(runInterval)
            }), 200)

        const progressInterval = setInterval(() => fetch(new Request(`/runs/${selectedRun}/results`))
            .then(resp => resp.json())
            .then(progress => this.updateProgress.bind(this)(selectedRun, progress))
            .catch(err => {
                console.error(err)
                clearInterval(progressInterval)
            }), 200)

        this.state = {
            // runInterval,
            progressInterval,
            totalMessages: runTotalMessages,
            progress: {},
        }
    }

    componentWillReceiveProps(nextProps) {
        const {selectedRun} = nextProps

        clearInterval(this.state.runInterval)
        clearInterval(this.state.progressInterval)

        const runInterval = setInterval(() => fetch(new Request(`/runs/${selectedRun}`))
            .then(resp => resp.json())
            .then(run => this.setState({run}))
            .catch(err => console.error(err)), 200)

        const progressInterval = setInterval(() => fetch(new Request(`/runs/${selectedRun}/results`))
            .then(resp => resp.json())
            .then(progress => this.updateProgress.bind(this)(selectedRun, progress))
            .catch(err => console.error(err)), 200)

        this.setState({
            runInterval,
            progressInterval,
            run: {},
            progress: {},
        })
    }

    componentWillUnmount() {
        clearInterval(this.state.runInterval)
        clearInterval(this.state.progressInterval)
    }

    componentWillUpdate(nextProps, nextState) {
        if (nextState.finishedCreating) {
            clearInterval(this.state.runInterval)
        }
    }

    render() {
        const {totalMessages, progress} = this.state

        const progressBars = Object.keys(progress)
            .filter(k => progress[k])
            .map(k => <div key={k}>
                <label>{k}</label>
                <progress value={progress[k]['count']} max={totalMessages}/>
            </div>)

        return <div className="results">
            <div>Sent messages: {totalMessages}</div>
            {progressBars}
        </div>
    }

    updateProgress(selectedRun, newProgress) {
        if (selectedRun !== this.props.selectedRun) {
            return
        }

        this.setState({
            progress: {
                'http': false,
                'batch-http': false,
                'udp': false,
                'quic': false,
                'websocket': false,
                'grpc-streaming': false,
                'grpc-unary': false,
                ...this.state.progress,
                ...newProgress,
            }
        })
    }
}

export default class ResultsArea extends React.Component {
    render() {
        const {selectedRun} = this.props

        let content = <div/>
        let subtitle = <small>No run selected</small>

        if (selectedRun) {
            content = <RunResults selectedRun={selectedRun}/>
            subtitle = <small>Viewing run {selectedRun}</small>
        }

        return <div className="results-area">
            <h3>Results area</h3>
            {subtitle}
            {content}
        </div>
    }
}