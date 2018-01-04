import '../css/app.scss' // required for wepback to build css

import React from 'react'
import {render} from 'react-dom'

import Sidebar from './components/Sidebar'
import ResultsArea from './components/ResultsArea'

render(<div className="page"><Sidebar/><ResultsArea/></div>, document.getElementById('root'))