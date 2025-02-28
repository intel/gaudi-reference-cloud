"use strict";

var dropdown = {};
var dropdown2 = {};

if(window.custom_allegria){

    document.addEventListener('DOMContentLoaded', function () {
        console.log('document is ready!');
        dropdown=returnDropDown("usecase");
        dropdown2=returnDropDown("devusecase");
        linkElementOpenInNewTab();
        if (dropdown) {
            createDropdown(useCases,dropdown);
            processorObjViewer(dropdown,useCases,'Processor', 'Use case');
            console.log("dropdown 1:",dropdown)
        }
        if (dropdown2) {
            createDropdown(devUseCases,dropdown2);
            processorObjViewer(dropdown2,devUseCases,'Processor', 'Use case');
            console.log("dropdown2 2:",dropdown2)
        };
    });
}

window.custom_allegria = ranDomizer();
 // todo: improve in sphinx-based solution; avoid dupe build

function ranDomizer() {
    let randomHash = (Math.random() + 1).toString(36).substring(2);
    return randomHash
}

function returnDropDown(id){
    const dropdownById = document.getElementById(id)
    return dropdownById
}

function linkElementOpenInNewTab() {
    const linkElement = document.querySelectorAll("a[href^='https://'], a[href^='http://']");
    linkElement.forEach(link => {
        link.setAttribute("target", "_blank")
        link.setAttribute("rel", "noopener")
    })
};

var useCases = [
    {
        name: 'Deep Learning Training',
        proc: [
            { type: 'Intel® Gaudi® 2 processor', desc: 'Use to accelerate model training on Intel Gaudi 2'},
            { type: 'Intel® Max Series GPU', desc: 'Use to accelerate model training'} ,
            { type: '5th Generation Intel® Xeon® Scalable Processor', desc: 'Use to efficiently perform model training' }
        ]
    },
    {
        name: 'Deep Learning Inference',
        proc: [
            { type: 'Intel® Gaudi® 2 processor', desc: 'Use with models optimized for Intel Gaudi 2 to get maximum performance'},
            { type: 'Intel® Max Series GPU', desc: 'Use to accelerate model performance'},
            { type: '4th Generation Intel® Xeon® Scalable Processor', desc: 'Use to efficiently perform model inference'}
        ]
    },
    {
        name: 'Data Analytics',
        proc: [
            { type: '4th Generation Intel® Xeon® Scalable Processor', desc: 'Use for end-to-end analytics workflows that require maximum versatility'}, 
            { type: '5th Generation Intel® Xeon® Scalable Processor', desc: 'Use to accelerate end-to-end analytics workflows that require maximum versatility'}
        ]
    },
    {
        name: 'Classical Machine Learning',
        proc: [
            { type: '4th Generation Intel® Xeon® Scalable Processor', desc: 'Use for end-to-end analytics workflows that require maximum versatility'},
            { type: '5th Generation Intel® Xeon® Scalable Processor', desc: 'Use to accelerate end-to-end analytics workflows that require maximum versatility'}
        ]
    },
    {
        name: 'HPC and Scientific Computing',
        proc: [
            { type: 'Intel® Max Series GPU', desc: 'Use to accelerate scalable workloads'},
            { type: '4th Generation Intel® Xeon® Scalable Processor', desc: 'Use to process scalable workloads'}, 
            { type: '5th Generation Intel® Xeon® Scalable Processor', desc: 'Use to process scalable workloads that require faster memory'}
        ]
    }
]

// devUseCases = independent records previously added are below.
//
// { type: '5th Generation Intel® Xeon® Scalable Processor', desc: 'Use for 7 Billion or fewer paramater models'},
// {
//     name: 'Classical Machine Learning',
//     proc: [
//         { type: 'Intel® Max Series GPU', desc: 'Use for traditional machine learning (e.g., numpy, xgboost, etc.)'},
//         { type: '5th Generation Intel® Xeon® Scalable Processor', desc: 'Use for traditional machine learning (e.g., numpy, xgboost, etc.)'}
//     ]
// }


var devUseCases = [
    {
        name: 'Pre-training LLMs',
        proc: [
            { type: 'Intel® Gaudi® 2 processor', desc: 'Use to accelerate model training on Intel Gaudi 2'}
        ]
    },
    {
        name: 'Inference for LLMs - General',
        proc: [
            { type: 'Intel® Gaudi® 2 processor', desc: 'Use for very large models with large batch sizes'},

            { type: 'Intel® Gaudi® 2 processor', desc: 'Use for Huggingface models'},

            { type: 'Intel® Max Series GPU', desc: 'Use for custom-built models'},

            { type: '5th Generation Intel® Xeon® Scalable Processor', desc: 'For models that rely on complex numbers like NeRF or FFTs.'} ]
    },
    {
        name: 'Fine tuning LLMs',
        proc: [
            { type: 'Intel® Gaudi® 2 processor', desc: 'Use for Huggingface models'},
            { type: 'Intel® Max Series GPU', desc: 'Use for custom-built models'}, 
        ]
    },
    {
        name: 'Inference for multiple LLMs or multi-model pipelines',
        proc: [
            { type: 'Intel® Max Series GPU', desc: 'Use for multiple, different models running on the same card'}
        ]
    }

]

function processorObjViewer(dropdownMenu,jsonObj, str1, str2) {
    const table = document.querySelector('table');
    const tbody = document.querySelector('tbody');
    dropdownMenu.addEventListener('change', function(event) {
    const headerRow = `
    <tr> 
    <th>${str1}</th>
    <th>${str2}</th>
    </tr>`;
    if(event.target.selectedIndex){

    let caption = table.createCaption();
    caption.textContent = "Recommendation";
    const procDetails = jsonObj[event.target.selectedIndex - 1].proc.map(p => {
        return `<td>${p.type}</td> <td>${p.desc}</td>`}).join('</tr><tr>');

    const completeTable = `${headerRow}<tr>${procDetails}</tr>`;
    tbody.innerHTML = `${completeTable}`
    }else{
        table.deleteCaption();
        tbody.innerHTML = headerRow
    }
    }
)};

function createDropdown(jsonObj, dropdownMenu) {
    for(let i =0; i<jsonObj.length; i++) {
        let option = document.createElement('option');
        option.text = jsonObj[i].name;
        dropdownMenu.appendChild(option);
    }
}




